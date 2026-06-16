// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	alertmodel "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildIssueResponseDeduplicatesAlertInstances(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	events := []alertmodel.AEventWithWRecord{
		newSystemHealthTestAlert("evt-1", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-20*time.Minute), now.Add(-10*time.Minute)),
		newSystemHealthTestAlert("evt-2", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-18*time.Minute), now.Add(-2*time.Minute)),
		newSystemHealthTestAlert("evt-3", "ErrorRateHigh", alertmodel.SeverityWarnLevel, alertmodel.StatusFiring, now.Add(-12*time.Minute), now.Add(-1*time.Minute)),
	}

	input := reportInput{
		key:                  "entry-order-service-post-orders",
		aggregationKey:       "entry::order-service::POST /orders",
		kind:                 issueKindActiveAlert,
		title:                "订单入口告警",
		scope:                "order-service / POST /orders",
		scopeKind:            "entry",
		startedAt:            formatEventTime(events[0].CreateTime),
		updatedAt:            formatEventTime(events[2].UpdateTime),
		availableCalcSources: []string{"alerts"},
		displaySources:       []string{sourceLabel("alerts")},
	}
	for _, event := range events {
		input.events = append(input.events, event)
		input.normalizedEvents = append(input.normalizedEvents, normalizeAlertEvent(event))
		input.alertNames = append(input.alertNames, event.Name)
		input.alertSeverities = append(input.alertSeverities, event.Severity)
		input.affectedObjects = uniqueAffectedObjects(append(input.affectedObjects, affectedObjectsFromEvent(event)...))
	}
	finalizeReportInput(&input, now)
	input.title = buildAggregatedIssueTitle(input)
	input.impact = buildRealImpact(input)

	scored := scoreIssues([]reportInput{input})
	if len(scored) != 1 {
		t.Fatalf("expected one scored issue, got %d", len(scored))
	}
	issues := buildIssues(scored)
	if len(issues) != 1 {
		t.Fatalf("expected one response issue, got %d", len(issues))
	}
	issue := issues[0]

	if issue.RawEventCount != 3 {
		t.Fatalf("expected raw event count 3, got %d", issue.RawEventCount)
	}
	if issue.DeduplicatedAlerts != 2 {
		t.Fatalf("expected 2 deduplicated alert instances, got %d", issue.DeduplicatedAlerts)
	}
	if issue.ActiveAlerts != 2 {
		t.Fatalf("expected 2 active deduplicated alerts, got %d", issue.ActiveAlerts)
	}
	if issue.AggregationKey != "entry::order-service::POST /orders" {
		t.Fatalf("unexpected aggregation key: %s", issue.AggregationKey)
	}
	if issue.Priority != "P0" {
		t.Fatalf("critical entry issue should be P0, got %s", issue.Priority)
	}
	assertStringSliceContains(t, issue.MergeReasons, "同一业务入口")
	assertStringSliceContains(t, issue.MergeReasons, "同一时间窗口")
	assertStringSliceContains(t, issue.MergeReasons, "多类告警共同指向")
	assertStringSliceContains(t, issue.MergeReasons, "重复触发已去重")
	if !strings.Contains(issue.MergeBasis, "3 条真实事件去重为 2 个告警实例") {
		t.Fatalf("merge basis does not explain deduplication: %s", issue.MergeBasis)
	}
	if issue.RiskBreakdown.EntryImpactScore <= 0 || issue.RiskBreakdown.AlertSeverityScore <= 0 || issue.RiskBreakdown.PathRisk <= 0 {
		t.Fatalf("risk breakdown should be populated, got %+v", issue.RiskBreakdown)
	}
	if issue.RiskBreakdown.PropagationScore <= 0 {
		t.Fatalf("duplicate and multi-alert entry should carry propagation pressure, got %+v", issue.RiskBreakdown)
	}
	assertAffectedObject(t, issue.AffectedObjects, "service", "order-service")
	assertAffectedObject(t, issue.AffectedObjects, "endpoint", "POST /orders")
}

func TestNormalizeAlertEventBuildsServiceEndpointAggregationKey(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	event := newSystemHealthTestAlert("evt-1", "HighLatency", "High", "active", now.Add(-5*time.Minute), now)

	normalized := normalizeAlertEvent(event)

	if normalized.Status != alertmodel.StatusFiring {
		t.Fatalf("expected active status to normalize to firing, got %s", normalized.Status)
	}
	if normalized.Severity != alertmodel.SeverityErrorLevel {
		t.Fatalf("expected High severity to normalize to error, got %s", normalized.Severity)
	}
	if normalized.AggregationKey != "entry::order-service::POST /orders" {
		t.Fatalf("unexpected aggregation key: %s", normalized.AggregationKey)
	}
	if normalized.DedupKey == "" {
		t.Fatal("expected dedup key to be populated")
	}
}

func TestBuildAggregatedIssueTitleUsesAlertSymptomNotObject(t *testing.T) {
	input := reportInput{
		kind:       issueKindActiveAlert,
		title:      "对象告警汇总：1 类告警",
		scope:      "ts-gateway-service / POST travel-plan",
		scopeKind:  "entry",
		alertNames: []string{"平均请求延时相比昨天升高50%以上"},
	}

	title := buildAggregatedIssueTitle(input)

	if title != "平均请求延时相比昨天升高50%以上" {
		t.Fatalf("expected alert symptom title, got %q", title)
	}
	if strings.Contains(title, input.scope) || strings.Contains(title, "告警汇总") {
		t.Fatalf("title should not fall back to object or aggregation wording: %q", title)
	}
}

func TestBuildAggregatedIssueTitleNamesMergedAlertTypes(t *testing.T) {
	input := reportInput{
		kind:       issueKindRecoveredAlert,
		title:      "对象告警汇总：2 类告警",
		scope:      "ts-gateway-service / POST travel-plan",
		scopeKind:  "entry",
		alertNames: []string{"平均请求延时相比昨天升高50%以上", "请求错误率相比昨天升高50%以上"},
	}

	title := buildAggregatedIssueTitle(input)

	if title != "已恢复告警：平均请求延时相比昨天升高50%以上等 2 类告警" {
		t.Fatalf("unexpected merged title: %q", title)
	}
}

func TestBuildSummaryUsesUserFacingPatternInsteadOfObjectList(t *testing.T) {
	scored := []scoredIssue{
		{
			input: reportInput{
				kind:                 issueKindRecoveredAlert,
				title:                "平均请求延时相比昨天升高50%以上",
				scope:                "ts-gateway-service / POST travel-plan",
				scopeKind:            "entry",
				rawEventCount:        6,
				deduplicatedAlerts:   2,
				activeAlerts:         0,
				alertNames:           []string{"平均请求延时相比昨天升高50%以上"},
				availableCalcSources: []string{"alerts"},
			},
		},
		{
			input: reportInput{
				kind:                 issueKindRecoveredAlert,
				title:                "平均请求延时相比昨天升高50%以上",
				scope:                "ts-order-service / POST order",
				scopeKind:            "entry",
				rawEventCount:        3,
				deduplicatedAlerts:   1,
				activeAlerts:         0,
				alertNames:           []string{"平均请求延时相比昨天升高50%以上"},
				availableCalcSources: []string{"alerts"},
			},
		},
	}

	summary := buildSummary(scored, 68)

	for _, expected := range []string{"系统当前处于警告状态", "窗口内发生过多入口请求延时告警", "反复触发", "恢复观察", "需要关注告警是否复发"} {
		if !strings.Contains(summary, expected) {
			t.Fatalf("expected summary to contain %q, got %q", expected, summary)
		}
	}
	for _, unexpected := range []string{"整体判断", "主要风险模式", "ts-gateway-service", "POST travel-plan", "优先处理故障对象"} {
		if strings.Contains(summary, unexpected) {
			t.Fatalf("summary should not enumerate object %q: %q", unexpected, summary)
		}
	}
}

func TestAlertManagerAlertToEventMatchesAlertgroupLabel(t *testing.T) {
	alert := alertManagerAPIAlert{
		Labels: map[string]string{
			"alertname":  "HighLatency",
			"alertgroup": "692489354112",
			"severity":   "critical",
			"namespace":  "apo",
			"pod":        "apo-one-agent-988hs",
		},
		Annotations: map[string]string{
			"summary": "latency is high",
		},
		StartsAt: "2026-06-03T01:49:00Z",
	}
	alert.Status.State = "active"

	event, ok := alertManagerAlertToEvent(alert, map[string]struct{}{"692489354112": {}})
	if !ok {
		t.Fatal("expected alertmanager alert to match by alertgroup label")
	}
	if event.Status != alertmodel.StatusFiring || !isValidAlert(event) {
		t.Fatalf("expected valid firing event, got status=%s validity=%s isValid=%s", event.Status, event.Validity, event.IsValid)
	}
	if event.Name != "HighLatency" || event.Group != "692489354112" {
		t.Fatalf("unexpected converted event: name=%s group=%s", event.Name, event.Group)
	}
}

func TestRecoveredAlertsRemainRecoveringAndLowImpact(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	event := newSystemHealthTestAlert("evt-1", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusResolved, now.Add(-30*time.Minute), now.Add(-20*time.Minute))
	input := reportInput{
		key:                  "recovered-entry-order-service-post-orders",
		aggregationKey:       "entry::order-service::POST /orders",
		kind:                 issueKindRecoveredAlert,
		title:                "已恢复订单入口告警",
		scope:                "order-service / POST /orders",
		scopeKind:            "entry",
		startedAt:            formatEventTime(event.CreateTime),
		updatedAt:            formatEventTime(event.UpdateTime),
		events:               []alertmodel.AEventWithWRecord{event},
		normalizedEvents:     []normalizedAlertEvent{normalizeAlertEvent(event)},
		alertNames:           []string{event.Name},
		alertSeverities:      []string{event.Severity},
		availableCalcSources: []string{"alerts"},
		displaySources:       []string{sourceLabel("alerts")},
		affectedObjects:      affectedObjectsFromEvent(event),
	}
	finalizeReportInput(&input, now)
	input.impact = buildRealImpact(input)

	scored := scoreIssues([]reportInput{input})
	issues := buildIssues(scored)
	if len(issues) != 1 {
		t.Fatalf("expected one recovered issue, got %d", len(issues))
	}
	issue := issues[0]
	if issue.Priority != "P3" {
		t.Fatalf("recovered alert should be P3, got %s", issue.Priority)
	}
	if issue.Status != "recovering" {
		t.Fatalf("recovered alert should be recovering, got %s", issue.Status)
	}
	if issue.ActiveAlerts != 0 {
		t.Fatalf("recovered alert should not count as active, got %d", issue.ActiveAlerts)
	}
	if issue.RiskBreakdown.EntryImpactScore >= 50 {
		t.Fatalf("recovered alert entry impact should be bounded, got %+v", issue.RiskBreakdown)
	}
}

func TestBuildInputsFromRealEventsKeepsIndependentIssuesSeparate(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	orderAlert := newSystemHealthTestAlert("evt-order", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-15*time.Minute), now.Add(-1*time.Minute))
	paymentAlert := newSystemHealthTestAlert("evt-payment", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-15*time.Minute), now.Add(-1*time.Minute))
	setAlertTags(&paymentAlert, map[string]string{
		"serviceName": "payment-service",
		"endpoint":    "POST /payments",
		"namespace":   "prod",
		"pod":         "payment-service-0",
	})

	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{orderAlert, paymentAlert}, now.Add(-30*time.Minute), now, telemetryCapabilityStatus{})
	if len(inputs) != 2 {
		t.Fatalf("expected independent service endpoints to produce 2 issues, got %d", len(inputs))
	}

	keys := make(map[string]reportInput, len(inputs))
	for _, input := range inputs {
		keys[input.aggregationKey] = input
		if input.rawEventCount != 1 || input.deduplicatedAlerts != 1 || input.activeAlerts != 1 {
			t.Fatalf("expected single alert instance for %s, got raw=%d dedup=%d active=%d", input.aggregationKey, input.rawEventCount, input.deduplicatedAlerts, input.activeAlerts)
		}
	}
	if _, ok := keys["entry::order-service::POST /orders"]; !ok {
		t.Fatalf("missing order issue in %#v", keys)
	}
	if _, ok := keys["entry::payment-service::POST /payments"]; !ok {
		t.Fatalf("missing payment issue in %#v", keys)
	}
}

func TestBuildInputsFromRealEventsKeepsInvalidAlertsAsP4Diagnostics(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	valid := newSystemHealthTestAlert("evt-valid", "HighLatency", alertmodel.SeverityWarnLevel, alertmodel.StatusFiring, now.Add(-10*time.Minute), now)
	invalid := newSystemHealthTestAlert("evt-invalid", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-10*time.Minute), now)
	invalid.Validity = "invalid"
	invalid.IsValid = "false"
	unknownStatus := newSystemHealthTestAlert("evt-unknown", "HighLatency", alertmodel.SeverityCriticalLevel, "pending", now.Add(-10*time.Minute), now)

	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{valid, invalid, unknownStatus}, now.Add(-30*time.Minute), now, telemetryCapabilityStatus{})
	if len(inputs) != 2 {
		t.Fatalf("expected valid firing issue and invalid diagnostic issue, got %d", len(inputs))
	}
	byKind := make(map[string]reportInput, len(inputs))
	for _, input := range inputs {
		byKind[input.kind] = input
	}
	if byKind[issueKindActiveAlert].events[0].EventID != "evt-valid" {
		t.Fatalf("expected valid event to remain as active alert, got %#v", byKind[issueKindActiveAlert].events)
	}
	invalidInput := byKind[issueKindInvalidAlert]
	if invalidInput.events[0].EventID != "evt-invalid" || !strings.Contains(invalidInput.title, "疑似告警规则问题") {
		t.Fatalf("expected invalid event to become diagnostic issue, got %#v", invalidInput)
	}
	scored := scoreIssues([]reportInput{invalidInput})
	if len(scored) != 1 || scored[0].priority != "P4" {
		t.Fatalf("expected invalid diagnostic to be P4, got %#v", scored)
	}
	plan := buildActionPlan(scored[0])
	if len(plan) == 0 || !strings.Contains(plan[0].Action, "RED 指标") {
		t.Fatalf("expected invalid diagnostic action plan to compare RED metrics, got %#v", plan)
	}
}

func TestObjectOnlyAlertsAggregateByObjectFallback(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	nodeAlert := newSystemHealthTestAlert("evt-node", "NodeDiskFull", alertmodel.SeverityErrorLevel, alertmodel.StatusFiring, now.Add(-45*time.Minute), now)
	setAlertTags(&nodeAlert, map[string]string{
		"node":      "node-a",
		"namespace": "infra",
	})
	nodeAlert.Group = "infra"

	normalized := normalizeAlertEvent(nodeAlert)
	if normalized.AggregationKey != "object::node::node-a" {
		t.Fatalf("expected node object aggregation key, got %s", normalized.AggregationKey)
	}

	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{nodeAlert}, now.Add(-60*time.Minute), now, telemetryCapabilityStatus{})
	if len(inputs) != 1 {
		t.Fatalf("expected one node issue, got %d", len(inputs))
	}
	issues := buildIssues(scoreIssues(inputs))
	if len(issues) != 1 {
		t.Fatalf("expected one node response issue, got %d", len(issues))
	}
	issue := issues[0]
	if issue.AggregationKey != "object::node::node-a" || issue.Scope != "node-a" {
		t.Fatalf("unexpected node issue aggregation: key=%s scope=%s", issue.AggregationKey, issue.Scope)
	}
	assertAffectedObject(t, issue.AffectedObjects, "node", "node-a")
}

func TestResponseDoesNotExposeInternalDataQualityScore(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	event := newSystemHealthTestAlert("evt-1", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-20*time.Minute), now)
	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{event}, now.Add(-30*time.Minute), now, telemetryCapabilityStatus{})
	issues := buildIssues(scoreIssues(inputs))
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}

	payload, err := json.Marshal(response.SystemHealthReportResp{
		Report: response.SystemHealthReport{Status: "severe", Score: 42},
		Issues: issues,
	})
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	for _, forbidden := range []string{"confidence", "Confidence", "dataQualityScore", "internalDataQualityScore"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("response should not expose %s: %s", forbidden, payload)
		}
	}
}

func TestBuildSystemHealthAIRequestIncludesExistingAlertAnalysis(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	event := newSystemHealthTestAlert("evt-1", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-20*time.Minute), now)
	event.AlertDirection = "疑似下游 payment-service 响应变慢导致入口延迟升高"
	event.AnalyzeRunID = "analyze-run-1"
	event.Output = "false"

	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{event}, now.Add(-30*time.Minute), now, telemetryCapabilityStatus{})
	inputs[0].alertReportActions = []systemHealthAlertReportAction{
		{AlertName: event.Name, Level: "high", Action: "先扩容 payment-service 实例并观察入口延迟是否恢复"},
	}
	scored := scoreIssues(inputs)
	issues := buildIssues(scored)
	if len(issues[0].ExistingAlertAnalyses) != 1 {
		t.Fatalf("expected issue response to expose existing alert analysis, got %#v", issues[0].ExistingAlertAnalyses)
	}
	if issues[0].ExistingAlertAnalyses[0].AlertDirection != event.AlertDirection {
		t.Fatalf("issue response did not preserve alert direction: %#v", issues[0].ExistingAlertAnalyses[0])
	}
	if !strings.Contains(issues[0].RootCause, "payment-service 响应变慢") {
		t.Fatalf("expected root cause to be organized from existing alert analysis, got %q", issues[0].RootCause)
	}
	resp := response.SystemHealthReportResp{
		Report: response.SystemHealthReport{
			Status:           "warning",
			Score:            68,
			ServiceRiskCount: len(issues),
			ActiveAlertCount: uniqueActiveAlertCount(scored),
			Summary:          buildSummary(scored, 68),
			Window:           "15m",
			AvailableSources: reportAvailableSources(scored),
		},
		Issues: issues,
	}

	aiReq := buildSystemHealthAIRequest(&resp, scored)
	if len(aiReq.Issues) != 1 {
		t.Fatalf("expected one AI issue input, got %d", len(aiReq.Issues))
	}
	issue := aiReq.Issues[0]
	if issue.Key != issues[0].Key || issue.Scope != "order-service / POST /orders" {
		t.Fatalf("unexpected issue input: %#v", issue)
	}
	if len(issue.ExistingAlertAnalyses) != 1 {
		t.Fatalf("expected existing alert analysis to be included, got %#v", issue.ExistingAlertAnalyses)
	}
	if len(issue.ExistingReportActions) != 1 || issue.ExistingReportActions[0].Action != "先扩容 payment-service 实例并观察入口延迟是否恢复" {
		t.Fatalf("expected original report action to be included, got %#v", issue.ExistingReportActions)
	}
	analysis := issue.ExistingAlertAnalyses[0]
	if analysis.AlertDirection != event.AlertDirection || analysis.AnalyzeRunID != "analyze-run-1" || analysis.CheckOutput != "false" {
		t.Fatalf("existing alert analysis was not preserved: %#v", analysis)
	}
	if len(issue.Evidence) == 0 {
		t.Fatalf("expected evidence summary in AI request: %#v", issue)
	}
}

func TestAlertReportSuggestActionsMergeIntoActionPlan(t *testing.T) {
	actions := parseAlertReportSuggestActions([]string{
		`{"action":"重启异常实例并观察错误率","level":"high"}`,
		`{"action":"重启异常实例并观察错误率","level":"high"}`,
		`{"action":"补充检查下游数据库连接池","level":"medium"}`,
	}, "HighLatency")
	if len(actions) != 2 {
		t.Fatalf("expected two unique report actions, got %#v", actions)
	}

	item := scoredIssue{
		input: reportInput{
			kind:               issueKindActiveAlert,
			scope:              "order-service / POST /orders",
			alertReportActions: actions,
		},
	}
	plan := buildActionPlan(item)
	if len(plan) < 3 {
		t.Fatalf("expected report actions plus default action plan, got %#v", plan)
	}
	if plan[0].Phase != "原报告高优先级建议" || plan[0].Action != "重启异常实例并观察错误率" {
		t.Fatalf("expected high priority report action first, got %#v", plan[0])
	}
	if plan[1].Phase != "原报告建议" || plan[1].Action != "补充检查下游数据库连接池" {
		t.Fatalf("expected medium report action second, got %#v", plan[1])
	}
}

func TestBuildIssuesIncludesStructuredActionPlanBeforeAIEnrichment(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	event := newSystemHealthTestAlert("evt-1", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-20*time.Minute), now)

	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{event}, now.Add(-30*time.Minute), now, telemetryCapabilityStatus{})
	issues := buildIssues(scoreIssues(inputs))
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	issue := issues[0]
	if issue.AIConclusion == "" || len(issue.Actions) == 0 || len(issue.ActionPlan) == 0 {
		t.Fatalf("expected structured AI fallback fields before direct AI enrichment, got conclusion=%q actions=%#v actionPlan=%#v", issue.AIConclusion, issue.Actions, issue.ActionPlan)
	}
}

func TestCallSystemHealthAIUsesDirectAIEndpoint(t *testing.T) {
	var gotRequest openAIResponsesRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("expected bearer auth header, got %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if gotRequest.Model != "gpt-5.5" {
			t.Fatalf("expected model gpt-5.5, got %s", gotRequest.Model)
		}
		if gotRequest.Text.Format.Type != "json_schema" || gotRequest.Text.Format.Name != "system_health_summary" || !gotRequest.Text.Format.Strict {
			t.Fatalf("expected strict json_schema response format, got %#v", gotRequest.Text.Format)
		}
		if gotRequest.Text.Format.Schema == nil {
			t.Fatal("expected response schema to be included")
		}
		if strings.Contains(gotRequest.Instructions, `{"summary"`) {
			t.Fatalf("instructions should not rely on an example JSON literal: %s", gotRequest.Instructions)
		}
		if !strings.Contains(gotRequest.Input, "payment-service 下游变慢") {
			t.Fatalf("expected existing alert analysis in OpenAI input: %s", gotRequest.Input)
		}
		aiOutput, err := json.Marshal(systemHealthAISummaryResponse{
			Summary: "当前 order-service 支付入口存在高优先级延迟问题，应先控制入口影响。",
			Issues: []systemHealthAIIssueOutput{
				{
					Key:       "issue-1",
					Analysis:  "order-service / POST /orders 延迟持续升高，已有告警分析指向 payment-service。",
					RootCause: "payment-service 下游响应变慢导致 order-service 入口延迟。",
					ActionPlan: []response.SystemHealthActionPlanStep{
						{
							Phase:      "影响控制",
							Target:     "order-service / POST /orders",
							Action:     "先确认入口错误率和延迟是否仍在升高。",
							Validation: "入口延迟恢复到阈值内。",
							Risk:       "若 payment-service 仍持续变慢，入口延迟可能反复升高。",
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("marshal AI output: %v", err)
		}
		_ = json.NewEncoder(w).Encode(openAIResponsesResponse{
			ID:         "resp-1",
			OutputText: string(aiOutput),
		})
	}))
	defer server.Close()

	payload := systemHealthAISummaryRequest{
		Window: "15m",
		Report: systemHealthAIReportInput{
			Status:  "warning",
			Score:   68,
			Summary: "结构化摘要",
		},
		Issues: []systemHealthAIIssueInput{
			{
				Key:      "issue-1",
				Priority: "P1",
				Scope:    "order-service / POST /orders",
				ExistingAlertAnalyses: []systemHealthAIAlertAnalysisInput{
					{AlertDirection: "payment-service 下游变慢", AnalyzeRunID: "analyze-run-1"},
				},
			},
		},
	}

	aiResp, err := callSystemHealthAI(nil, server.URL, "test-key", "gpt-5.5", 3, payload)
	if err != nil {
		t.Fatalf("call AI summary API: %v", err)
	}
	if aiResp.Summary == "" || aiResp.WorkflowRunID != "resp-1" || len(aiResp.Issues) != 1 {
		t.Fatalf("unexpected AI response: %#v", aiResp)
	}
	if aiResp.Issues[0].RootCause != "payment-service 下游响应变慢导致 order-service 入口延迟。" {
		t.Fatalf("expected rootCause from AI response, got %#v", aiResp.Issues[0])
	}
}

func TestCallSystemHealthAIIntegrationUsesRealCompanyResponsesAPI(t *testing.T) {
	if os.Getenv("SYSTEM_HEALTH_AI_INTEGRATION") != "1" {
		t.Skip("set SYSTEM_HEALTH_AI_INTEGRATION=1 to call the real company Responses API")
	}
	apiKey := strings.TrimSpace(os.Getenv("COMPANY_API_KEY"))
	if apiKey == "" {
		t.Fatal("COMPANY_API_KEY is required for real AI integration verification")
	}
	baseURL := firstNonEmpty(os.Getenv("SYSTEM_HEALTH_AI_BASE_URL"), "http://192.168.1.8:8319/v1")
	model := firstNonEmpty(os.Getenv("SYSTEM_HEALTH_AI_MODEL"), "gpt-5.5")

	payload := systemHealthAISummaryRequest{
		Window: "15m",
		Report: systemHealthAIReportInput{
			Status:           "warning",
			Score:            68,
			ServiceRiskCount: 1,
			ActiveAlertCount: 2,
			AvailableSources: []string{"真实告警事件", "已有告警分析产出", "Prometheus指标"},
			Summary:          "当前系统健康度为 68 分；order-service / POST /orders 存在高优先级延迟问题。",
		},
		Issues: []systemHealthAIIssueInput{
			{
				Key:          "real-ai-verification-issue",
				Priority:     "P1",
				Title:        "订单入口延迟升高",
				Status:       "active",
				Scope:        "order-service / POST /orders",
				Impact:       "订单创建入口延迟升高，可能影响下单体验和支付链路。",
				ActiveAlerts: 2,
				StartedAt:    "2026-06-16 10:30",
				DataSources:  []string{"真实告警事件", "已有告警分析产出", "Prometheus指标"},
				Evidence: []systemHealthAIEvidenceInput{
					{Type: "告警指标趋势", Source: "Prometheus", Severity: "critical", Result: "HighLatency 持续 firing，入口 p95 延迟高于阈值。"},
					{Type: "已有告警分析", Source: "alert-analysis", Severity: "critical", Result: "已有分析指向 payment-service 下游响应变慢。"},
				},
				ExistingAlertAnalyses: []systemHealthAIAlertAnalysisInput{
					{
						EventID:        "real-ai-verification-event",
						AlertID:        "alert-HighLatency",
						AlertName:      "HighLatency",
						AlertDirection: "疑似下游 payment-service 响应变慢导致 order-service 入口延迟升高",
						AnalyzeRunID:   "real-ai-verification-run",
						CheckOutput:    "valid",
						Validity:       "valid",
					},
				},
			},
		},
	}

	aiResp, err := callSystemHealthAI(nil, baseURL, apiKey, model, 30, payload)
	if err != nil {
		t.Fatalf("real AI call failed: %v", err)
	}
	if strings.TrimSpace(aiResp.WorkflowRunID) == "" {
		t.Fatalf("expected real response id to be captured, got %#v", aiResp)
	}
	if strings.TrimSpace(aiResp.Summary) == "" || strings.Contains(aiResp.Summary, "置信度") {
		t.Fatalf("unexpected real AI summary: %q", aiResp.Summary)
	}
	if len(aiResp.Issues) != 1 {
		t.Fatalf("expected one issue in real AI response, got %#v", aiResp.Issues)
	}
	issue := aiResp.Issues[0]
	if issue.Key != "real-ai-verification-issue" {
		t.Fatalf("real AI response should preserve issue key, got %q", issue.Key)
	}
	if strings.TrimSpace(issue.Analysis) == "" || strings.Contains(issue.Analysis, "置信度") {
		t.Fatalf("unexpected real AI issue analysis: %q", issue.Analysis)
	}
	if len(issue.ActionPlan) == 0 {
		t.Fatalf("expected real AI action plan, got %#v", issue)
	}
	firstStep := issue.ActionPlan[0]
	if strings.TrimSpace(firstStep.Phase) == "" || strings.TrimSpace(firstStep.Target) == "" || strings.TrimSpace(firstStep.Action) == "" || strings.TrimSpace(firstStep.Validation) == "" || strings.TrimSpace(firstStep.Risk) == "" {
		t.Fatalf("real AI action step should contain all required fields, got %#v", firstStep)
	}
	t.Logf("real AI response id=%s summary=%s firstAction=%s", aiResp.WorkflowRunID, truncateText(aiResp.Summary, 120), truncateText(firstStep.Action, 120))
}

func TestCallSystemHealthAIReturnsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := callSystemHealthAI(nil, server.URL, "test-key", "gpt-5.5", 3, systemHealthAISummaryRequest{})
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("expected AI HTTP error, got %v", err)
	}
}

func TestResolveSystemHealthAIAPIKeyUsesEnvKey(t *testing.T) {
	t.Setenv("COMPANY_API_KEY", "resolved-company-key")

	if got := resolveSystemHealthAIAPIKey("", "COMPANY_API_KEY"); got != "resolved-company-key" {
		t.Fatalf("expected API key from env, got %q", got)
	}
	if got := resolveSystemHealthAIAPIKey("inline-key", "COMPANY_API_KEY"); got != "inline-key" {
		t.Fatalf("expected inline API key to take precedence, got %q", got)
	}
}

func TestBuildAlertEventEvidenceSamplesIncludesEventLabels(t *testing.T) {
	event := newSystemHealthTestAlert(
		"evt-1",
		"HighLatency",
		alertmodel.SeverityCriticalLevel,
		alertmodel.StatusFiring,
		time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC),
		time.Date(2026, 6, 16, 10, 35, 0, 0, time.UTC),
	)
	event.Detail = `{"summary":"Latency high VALUE = 1.23\nLABELS = map[serviceName:order-service endpoint:POST /orders namespace:prod]"}`

	samples := buildAlertEventEvidenceSamples(reportInput{
		events: []alertmodel.AEventWithWRecord{event},
	})

	if len(samples) != 1 {
		t.Fatalf("expected 1 event evidence sample, got %d", len(samples))
	}
	sample := samples[0]
	if sample.EventID != "evt-1" || sample.AlertID != "alert-HighLatency" {
		t.Fatalf("unexpected event identity: eventId=%q alertId=%q", sample.EventID, sample.AlertID)
	}
	if sample.Name != "HighLatency" || sample.Status != alertmodel.StatusFiring || sample.Severity != alertmodel.SeverityCriticalLevel {
		t.Fatalf("unexpected event metadata: %#v", sample)
	}
	if sample.StartedAt != "2026-06-16 10:30" {
		t.Fatalf("unexpected startedAt: %s", sample.StartedAt)
	}
	if sample.Labels["serviceName"] != "order-service" || sample.Labels["endpoint"] != "POST /orders" || sample.Labels["namespace"] != "prod" {
		t.Fatalf("expected selected alert labels, got %#v", sample.Labels)
	}
	if !strings.Contains(sample.Summary, "Latency high") || !strings.Contains(sample.Summary, "触发值=1.23") {
		t.Fatalf("expected alert detail summary with trigger value, got %q", sample.Summary)
	}
}

func TestRelatedObjectStatusCandidatesExtractsKubernetesLabels(t *testing.T) {
	event := newSystemHealthTestAlert(
		"evt-1",
		"HighLatency",
		alertmodel.SeverityCriticalLevel,
		alertmodel.StatusFiring,
		time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC),
		time.Date(2026, 6, 16, 10, 35, 0, 0, time.UTC),
	)
	event.EnrichTags["container"] = "app"
	event.EnrichTags["node_name"] = "worker-1"

	candidates := relatedObjectStatusCandidates([]alertmodel.AEventWithWRecord{event})
	if len(candidates) != 1 {
		t.Fatalf("expected one status candidate, got %#v", candidates)
	}
	candidate := candidates[0]
	if candidate.Namespace != "prod" || candidate.Pod != "order-service-0" || candidate.Container != "app" || candidate.Node != "worker-1" {
		t.Fatalf("unexpected status candidate: %#v", candidate)
	}
}

func TestPodObjectStatusIncludesContainerReadinessAndRestarts(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "order-service-0",
			Namespace:         "prod",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-10 * time.Minute)),
			Labels: map[string]string{
				"app":               "order-service",
				"pod-template-hash": "abc123",
				"ignored":           "value",
			},
		},
		Spec: corev1.PodSpec{
			NodeName: "worker-1",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "app",
					Ready:        true,
					RestartCount: 2,
					Image:        "order-service:v1",
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
				{
					Name:         "sidecar",
					Ready:        false,
					RestartCount: 1,
					Image:        "sidecar:v1",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
					},
				},
			},
		},
	}

	status := podObjectStatus(pod, "sidecar")
	if status.Type != "pod" || status.Name != "order-service-0" || status.Namespace != "prod" || status.Status != "Running" {
		t.Fatalf("unexpected pod status identity: %#v", status)
	}
	if status.Ready == nil || !*status.Ready {
		t.Fatalf("expected pod ready=true, got %#v", status.Ready)
	}
	if status.Restarts != 3 || status.NodeName != "worker-1" {
		t.Fatalf("unexpected pod restart/node status: %#v", status)
	}
	if len(status.Containers) != 1 || status.Containers[0].Name != "sidecar" || status.Containers[0].State != "waiting" || status.Containers[0].Reason != "CrashLoopBackOff" {
		t.Fatalf("expected selected sidecar status, got %#v", status.Containers)
	}
	if status.Labels["app"] != "order-service" || status.Labels["ignored"] != "" {
		t.Fatalf("expected selected labels only, got %#v", status.Labels)
	}
}

func TestMetricEvidenceMarksEventScopedSeriesMatch(t *testing.T) {
	event := newSystemHealthTestAlert(
		"evt-1",
		"HighLatency",
		alertmodel.SeverityCriticalLevel,
		alertmodel.StatusFiring,
		time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC),
		time.Date(2026, 6, 16, 10, 35, 0, 0, time.UTC),
	)
	results := []alertMetricSeries{
		{
			Labels: map[string]string{
				"serviceName": "order-service",
				"endpoint":    "POST /orders",
				"namespace":   "prod",
			},
			Values: []prometheus.Points{
				{TimeStamp: 1, Value: 1},
				{TimeStamp: 2, Value: 2},
			},
		},
		{
			Labels: map[string]string{
				"serviceName": "catalog",
				"endpoint":    "GET /items",
			},
			Values: []prometheus.Points{
				{TimeStamp: 1, Value: 10},
				{TimeStamp: 2, Value: 20},
			},
		},
	}

	matched := filterSeriesByAlertEvent(results, event)
	if len(matched) != 1 {
		t.Fatalf("expected one event-scoped series, got %d", len(matched))
	}
	stats, ok := calculateMetricTrendStats(matched)
	if !ok {
		t.Fatal("expected metric trend stats")
	}
	metric := buildMetricEvidence(alertMetricTrendQueryResult{
		Rule:           &request.AlertRule{Group: "app", Alert: "HighLatency"},
		Expression:     "sum by(serviceName,endpoint) (rate(http_request_duration_seconds_sum[5m]))",
		Step:           time.Minute,
		SeriesReturned: len(results),
		MatchedSeries:  matched,
		Stats:          stats,
		MatchLabels:    selectedAlertLabelMap(rawTagsFromEvent(event), 8),
		EventScoped:    true,
		Message:        "已按告警事件标签命中对应指标序列。",
	})

	if !metric.EventScoped {
		t.Fatal("expected metric evidence to be event scoped")
	}
	if metric.SeriesReturned != 2 || metric.SeriesMatched != 1 || metric.PointCount != 2 {
		t.Fatalf("unexpected metric counters: %#v", metric)
	}
	if metric.FirstValue != 1 || metric.LastValue != 2 || metric.MinValue != 1 || metric.MaxValue != 2 || metric.Delta != 1 {
		t.Fatalf("unexpected metric values: %#v", metric)
	}
	if metric.ChangePercent == nil || *metric.ChangePercent != 100 {
		t.Fatalf("expected 100%% change, got %#v", metric.ChangePercent)
	}
	if metric.MatchedLabels["serviceName"] != "order-service" || metric.MatchedLabels["endpoint"] != "POST /orders" {
		t.Fatalf("expected matched event labels, got %#v", metric.MatchedLabels)
	}
	if !strings.Contains(metric.MatchExpression, `serviceName="order-service"`) || !strings.Contains(metric.MatchExpression, `endpoint="POST /orders"`) {
		t.Fatalf("unexpected match expression: %s", metric.MatchExpression)
	}
	if len(metric.Series) != 1 || len(metric.Series[0].Points) != 2 {
		t.Fatalf("expected matched series points for charting, got %#v", metric.Series)
	}
	if metric.Series[0].Points[0].Timestamp != 1 || metric.Series[0].Points[0].Value != 1 {
		t.Fatalf("unexpected first chart point: %#v", metric.Series[0].Points[0])
	}
}

func TestSystemHealthReportCacheExpiresEntries(t *testing.T) {
	cache := newSystemHealthReportCache()
	key := "group=1:start=100:end=200"
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	report := response.SystemHealthReportResp{
		Report: response.SystemHealthReport{
			Status: "warning",
			Score:  68,
		},
	}

	cache.set(key, report, now)
	cached, ok := cache.get(key, now.Add(systemHealthReportCacheTTL-time.Second))
	if !ok {
		t.Fatal("expected report cache hit before expiry")
	}
	if cached.Report.Score != 68 || cached.Report.Status != "warning" {
		t.Fatalf("unexpected cached report: %#v", cached.Report)
	}

	if _, ok := cache.get(key, now.Add(systemHealthReportCacheTTL+time.Second)); ok {
		t.Fatal("expected report cache miss after expiry")
	}
}

func TestResolveTimeRangeUsesRelativeWindowWithoutCacheBucket(t *testing.T) {
	req := &request.GetSystemHealthReportRequest{RelativeTime: "15m"}

	startTime, endTime, window := resolveTimeRange(req)

	if window != "15m" {
		t.Fatalf("expected relative window label 15m, got %s", window)
	}
	if endTime.Sub(startTime) != 15*time.Minute {
		t.Fatalf("expected 15m window, got %s", endTime.Sub(startTime))
	}
}

func TestMetricEvidenceMarksGlobalBoundaryWhenSeriesCannotMatchEvent(t *testing.T) {
	results := []alertMetricSeries{
		{
			Labels: map[string]string{},
			Values: []prometheus.Points{
				{TimeStamp: 1, Value: 10},
				{TimeStamp: 2, Value: 15},
			},
		},
	}
	stats, ok := calculateMetricTrendStats(results)
	if !ok {
		t.Fatal("expected global metric trend stats")
	}

	metric := buildMetricEvidence(alertMetricTrendQueryResult{
		Rule:           &request.AlertRule{Group: "global", Alert: "GlobalHighLatency"},
		Expression:     "sum(rate(http_request_duration_seconds_sum[5m]))",
		Step:           2 * time.Minute,
		SeriesReturned: len(results),
		Stats:          stats,
		EventScoped:    false,
		Message:        "规则查询只返回全局或不可匹配的聚合序列，未能用当前告警事件标签定位到指定 series。",
	})

	if metric.EventScoped {
		t.Fatal("expected global metric evidence to be marked as not event scoped")
	}
	if metric.MatchExpression != "" || len(metric.MatchedLabels) != 0 {
		t.Fatalf("expected no event match labels, got expr=%q labels=%#v", metric.MatchExpression, metric.MatchedLabels)
	}
	if !strings.Contains(metric.Message, "未能用当前告警事件标签定位") {
		t.Fatalf("expected boundary message, got %q", metric.Message)
	}
	if len(metric.Series) != 0 {
		t.Fatalf("global boundary metric should not expose chart series, got %#v", metric.Series)
	}
}

func TestSystemHealthEvidenceJSONCarriesStructuredEvidence(t *testing.T) {
	change := 12.5
	evidence := response.SystemHealthEvidence{
		Type:   "告警指标趋势",
		Source: "真实告警事件/Prometheus",
		Query:  "sum(rate(http_requests_total[5m]))",
		Result: "已匹配告警规则。",
		AlertEvents: []response.SystemHealthAlertEventEvidence{
			{
				EventID:  "evt-1",
				AlertID:  "alert-HighLatency",
				Name:     "HighLatency",
				Status:   alertmodel.StatusFiring,
				Severity: alertmodel.SeverityCriticalLevel,
				Labels: map[string]string{
					"serviceName": "order-service",
				},
			},
		},
		Metric: &response.SystemHealthMetricEvidence{
			RuleGroup:      "app",
			RuleName:       "HighLatency",
			Expression:     "sum(rate(http_requests_total[5m]))",
			SeriesReturned: 2,
			SeriesMatched:  1,
			ChangePercent:  &change,
			EventScoped:    true,
		},
	}

	payload, err := json.Marshal(evidence)
	if err != nil {
		t.Fatalf("marshal evidence: %v", err)
	}
	jsonText := string(payload)
	for _, expected := range []string{`"alertEvents"`, `"metric"`, `"eventScoped":true`, `"seriesMatched":1`, `"serviceName":"order-service"`} {
		if !strings.Contains(jsonText, expected) {
			t.Fatalf("expected %s in json payload: %s", expected, jsonText)
		}
	}
}

func TestBuildAlertEventEvidenceResultIncludesEventPattern(t *testing.T) {
	now := time.Now()
	first := newSystemHealthTestAlert("evt-1", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusResolved, now.Add(-20*time.Minute), now.Add(-10*time.Minute))
	second := newSystemHealthTestAlert("evt-2", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusResolved, now.Add(-18*time.Minute), now.Add(-5*time.Minute))
	third := newSystemHealthTestAlert("evt-3", "ErrorRateHigh", alertmodel.SeverityWarnLevel, alertmodel.StatusResolved, now.Add(-16*time.Minute), now.Add(-4*time.Minute))
	input := reportInput{
		kind:   issueKindRecoveredAlert,
		scope:  "order-service / POST /orders",
		events: []alertmodel.AEventWithWRecord{first, second, third},
	}
	for _, event := range input.events {
		normalized := normalizeAlertEvent(event)
		input.normalizedEvents = append(input.normalizedEvents, normalized)
		input.alertNames = append(input.alertNames, event.Name)
		input.alertSeverities = append(input.alertSeverities, normalized.Severity)
	}
	finalizeReportInput(&input, now)

	result := buildAlertEventEvidenceResult(input)
	for _, expected := range []string{"事件规律=", "反复出现", "2类告警", "2个去重实例"} {
		if !strings.Contains(result, expected) {
			t.Fatalf("expected %q in evidence result: %s", expected, result)
		}
	}
}

func newSystemHealthTestAlert(eventID string, name string, severity string, status string, createTime time.Time, updateTime time.Time) alertmodel.AEventWithWRecord {
	tags := map[string]string{
		"serviceName": "order-service",
		"endpoint":    "POST /orders",
		"namespace":   "prod",
		"pod":         "order-service-0",
	}
	rawTags := alertmodel.RawTags{}
	for key, value := range tags {
		rawTags[key] = value
	}
	return alertmodel.AEventWithWRecord{
		AlertEvent: alertmodel.AlertEvent{
			Alert: alertmodel.Alert{
				Source:     "prometheus",
				SourceID:   "prometheus-main",
				AlertID:    "alert-" + name,
				Group:      "app",
				Name:       name,
				EnrichTags: tags,
				Tags:       rawTags,
			},
			EventID:      eventID,
			Detail:       `{"summary":"test alert"}`,
			CreateTime:   createTime,
			UpdateTime:   updateTime,
			ReceivedTime: updateTime,
			Severity:     severity,
			Status:       status,
		},
		WorkflowDetail: alertmodel.WorkflowDetail{
			Validity: "valid",
			IsValid:  "true",
		},
	}
}

func setAlertTags(event *alertmodel.AEventWithWRecord, tags map[string]string) {
	rawTags := alertmodel.RawTags{}
	for key, value := range tags {
		rawTags[key] = value
	}
	event.EnrichTags = tags
	event.Tags = rawTags
}

func assertStringSliceContains(t *testing.T, values []string, expected string) {
	t.Helper()
	for _, value := range values {
		if value == expected {
			return
		}
	}
	t.Fatalf("expected %q in %#v", expected, values)
}

func assertAffectedObject(t *testing.T, objects []response.SystemHealthAffectedObject, objectType string, name string) {
	t.Helper()
	for _, object := range objects {
		if object.Type == objectType && object.Name == name {
			return
		}
	}
	t.Fatalf("expected affected object %s:%s in %#v", objectType, name, objects)
}
