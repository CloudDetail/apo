// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	alertmodel "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
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
	assertStringSliceContains(t, issue.MergeReasons, "同一入口侧接口")
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

func TestNormalizeAlertEventBuildsServiceAggregationKey(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	event := newSystemHealthTestAlert("evt-1", "HighLatency", "High", "active", now.Add(-5*time.Minute), now)

	normalized := normalizeAlertEvent(event)

	if normalized.Status != alertmodel.StatusFiring {
		t.Fatalf("expected active status to normalize to firing, got %s", normalized.Status)
	}
	if normalized.Severity != alertmodel.SeverityErrorLevel {
		t.Fatalf("expected High severity to normalize to error, got %s", normalized.Severity)
	}
	if normalized.AggregationKey != "service::order-service" {
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

	if title != "ts-gateway-service / POST travel-plan 请求延时上升" {
		t.Fatalf("expected alert symptom title, got %q", title)
	}
	if !strings.Contains(title, "ts-gateway-service") || !strings.Contains(title, "请求延时上升") || strings.Contains(title, "告警汇总") {
		t.Fatalf("title should use object + problem style: %q", title)
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

	if !strings.HasPrefix(title, "已恢复：") || !strings.Contains(title, "ts-gateway-service") || !strings.Contains(title, "请求延时上升") {
		t.Fatalf("unexpected merged title: %q", title)
	}
}

func TestAcceptableAIIssueTitleRejectsUnsafeTitles(t *testing.T) {
	if !isAcceptableAIIssueTitle("order-service 延迟和错误率风险") {
		t.Fatal("expected concise AI issue title to be accepted")
	}
	if !isAcceptableAIIssueTitle("已恢复：order-service 存在请求延时问题") {
		t.Fatal("expected recovered prefix AI issue title to be accepted")
	}
	for _, title := range []string{
		"",
		"abc",
		"order-service\n延迟风险",
		"sum by(svc_name)(rate(http_requests_total[5m]))",
		"ts-station-service请求延时恢复观察",
		"order-service 延迟已恢复",
		"payment-service 触发中错误率风险",
		strings.Repeat("服务", 25),
	} {
		if isAcceptableAIIssueTitle(title) {
			t.Fatalf("expected unsafe AI issue title to be rejected: %q", title)
		}
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

func TestBuildInputsFromRealEventsKeepsIndependentServicesSeparate(t *testing.T) {
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
	if _, ok := keys["service::order-service"]; !ok {
		t.Fatalf("missing order issue in %#v", keys)
	}
	if _, ok := keys["service::payment-service"]; !ok {
		t.Fatalf("missing payment issue in %#v", keys)
	}
}

func TestBuildInputsFromRealEventsMergesSameServiceEndpoints(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	tripsAlert := newSystemHealthTestAlert("evt-trips", "HighLatency", alertmodel.SeverityWarnLevel, alertmodel.StatusResolved, now.Add(-20*time.Minute), now.Add(-10*time.Minute))
	routesAlert := newSystemHealthTestAlert("evt-routes", "HighLatency", alertmodel.SeverityWarnLevel, alertmodel.StatusResolved, now.Add(-18*time.Minute), now.Add(-8*time.Minute))
	setAlertTags(&routesAlert, map[string]string{
		"serviceName": "order-service",
		"endpoint":    "GET /routes/{tripId}",
		"namespace":   "prod",
		"pod":         "order-service-1",
	})

	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{tripsAlert, routesAlert}, now.Add(-30*time.Minute), now, telemetryCapabilityStatus{})
	if len(inputs) != 1 {
		t.Fatalf("expected same service endpoints to merge into one issue, got %d", len(inputs))
	}
	input := inputs[0]
	if input.aggregationKey != "service::order-service" || input.scope != "order-service" || input.scopeKind != "service" {
		t.Fatalf("expected service-level aggregation, got key=%s scope=%s scopeKind=%s", input.aggregationKey, input.scope, input.scopeKind)
	}
	if input.rawEventCount != 2 || input.deduplicatedAlerts != 2 {
		t.Fatalf("expected both endpoint alerts to remain represented, got raw=%d dedup=%d", input.rawEventCount, input.deduplicatedAlerts)
	}
	assertStringSliceContains(t, input.mergeReasons, "同一服务")
	assertAffectedObject(t, input.affectedObjects, "endpoint", "POST /orders")
	assertAffectedObject(t, input.affectedObjects, "endpoint", "GET /routes/{tripId}")
}

func TestApplyAlertReportTopologyRolePreservesServiceAggregation(t *testing.T) {
	input := reportInput{
		key:               "recovered-service-order-service",
		aggregationKey:    "service::order-service",
		kind:              issueKindRecoveredAlert,
		scope:             "order-service",
		scopeKind:         "service",
		objectCriticality: 80,
		alertReportTopologyRoles: []systemHealthAlertReportTopologyRole{
			{
				Role:            "internalEndpoint",
				Reason:          "topology_current_has_upstream_parents",
				CurrentService:  "order-service",
				CurrentEndpoint: "POST /orders",
				ParentCount:     1,
			},
		},
	}
	input.mergeReasons = buildMergeReasons(input)

	applyAlertReportTopologyRoleToInput(nil, &input)

	if input.aggregationKey != "service::order-service" || input.scope != "order-service" || input.scopeKind != "service" {
		t.Fatalf("topology role should not downgrade service aggregation, got key=%s scope=%s scopeKind=%s", input.aggregationKey, input.scope, input.scopeKind)
	}
	assertStringSliceContains(t, input.mergeReasons, "同一服务")
	if !strings.Contains(strings.Join(input.mergeReasons, "、"), "上游调用方") {
		t.Fatalf("expected topology context to remain in merge reasons, got %#v", input.mergeReasons)
	}
}

func TestBuildInputsFromRealEventsMergesInvalidDiagnosticsIntoSameScopeRisk(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	valid := newSystemHealthTestAlert("evt-valid", "HighLatency", alertmodel.SeverityWarnLevel, alertmodel.StatusFiring, now.Add(-10*time.Minute), now)
	invalid := newSystemHealthTestAlert("evt-invalid", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-10*time.Minute), now)
	invalid.Validity = "invalid"
	invalid.IsValid = "false"
	unknownStatus := newSystemHealthTestAlert("evt-unknown", "HighLatency", alertmodel.SeverityCriticalLevel, "pending", now.Add(-10*time.Minute), now)

	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{valid, invalid, unknownStatus}, now.Add(-30*time.Minute), now, telemetryCapabilityStatus{})
	if len(inputs) != 1 {
		t.Fatalf("expected same-scope valid alert and invalid diagnostic to merge into one issue, got %d", len(inputs))
	}
	input := inputs[0]
	if input.kind != issueKindActiveAlert {
		t.Fatalf("expected valid active alert to drive merged issue kind, got %s", input.kind)
	}
	if input.rawEventCount != 2 || input.activeAlerts != 1 || input.invalidAlerts != 1 {
		t.Fatalf("expected merged issue to keep valid and diagnostic counts, got raw=%d active=%d invalid=%d", input.rawEventCount, input.activeAlerts, input.invalidAlerts)
	}
	assertStringSliceContains(t, input.mergeReasons, "包含告警规则诊断")
	if strings.Contains(input.title, "疑似告警规则问题") {
		t.Fatalf("merged service risk title should not be dominated by rule diagnostic wording: %q", input.title)
	}
	scored := scoreIssues(inputs)
	if len(scored) != 1 || scored[0].priority == "P4" {
		t.Fatalf("expected merged valid risk not to degrade into standalone P4 diagnostic, got %#v", scored)
	}
	plan := buildActionPlan(scored[0])
	foundRuleCheck := false
	for _, step := range plan {
		if strings.Contains(step.Phase, "规则") && strings.Contains(step.Action, "RED 指标") {
			foundRuleCheck = true
			break
		}
	}
	if !foundRuleCheck {
		t.Fatalf("expected merged action plan to retain rule diagnostic check, got %#v", plan)
	}
}

func TestBuildInputsFromRealEventsKeepsPureInvalidAlertsAsP4Diagnostics(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	invalid := newSystemHealthTestAlert("evt-invalid", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-10*time.Minute), now)
	invalid.Validity = "invalid"
	invalid.IsValid = "false"

	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{invalid}, now.Add(-30*time.Minute), now, telemetryCapabilityStatus{})
	if len(inputs) != 1 {
		t.Fatalf("expected one invalid diagnostic issue, got %d", len(inputs))
	}
	input := inputs[0]
	if input.kind != issueKindInvalidAlert || input.invalidAlerts != 1 || !strings.Contains(input.title, "存在告警规则问题") {
		t.Fatalf("expected pure invalid event to remain standalone diagnostic issue, got %#v", input)
	}
	scored := scoreIssues(inputs)
	if len(scored) != 1 || scored[0].priority != "P4" {
		t.Fatalf("expected pure invalid diagnostic to be P4, got %#v", scored)
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
	inputs[0].alertReportFindings = []systemHealthAlertReportFinding{
		{AlertName: event.Name, RootCause: "payment-service 下游响应变慢导致入口延迟升高", AnalyzeProcess: "报告定位到 payment-service 调用耗时升高。", Detail: "overview_detail 显示 payment-service 数据库查询耗时升高，入口等待下游响应导致订单接口延迟。"},
	}
	scored := scoreIssues(inputs)
	issues := buildIssues(scored)
	if len(issues[0].ExistingAlertAnalyses) != 1 {
		t.Fatalf("expected issue response to expose existing alert analysis, got %#v", issues[0].ExistingAlertAnalyses)
	}
	if issues[0].ExistingAlertAnalyses[0].AlertDirection != event.AlertDirection {
		t.Fatalf("issue response did not preserve alert direction: %#v", issues[0].ExistingAlertAnalyses[0])
	}
	if !strings.Contains(issues[0].RootCause, "payment-service 下游响应变慢") {
		t.Fatalf("expected root cause to be organized from alert report finding, got %q", issues[0].RootCause)
	}
	if !strings.Contains(issues[0].RootCause, "告警分析报告判断") {
		t.Fatalf("expected root cause to prefer alert report finding, got %q", issues[0].RootCause)
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
	if issue.Key != issues[0].Key || issue.Scope != "order-service" {
		t.Fatalf("unexpected issue input: %#v", issue)
	}
	if len(issue.AlertNames) != 1 || issue.AlertNames[0] != "HighLatency" {
		t.Fatalf("expected real alert names to be included for title synthesis, got %#v", issue.AlertNames)
	}
	if len(issue.ExistingAlertAnalyses) != 1 {
		t.Fatalf("expected existing alert analysis to be included, got %#v", issue.ExistingAlertAnalyses)
	}
	if len(issue.ExistingReportFindings) != 1 || !strings.Contains(issue.ExistingReportFindings[0].RootCause, "payment-service") {
		t.Fatalf("expected alert report finding to be included, got %#v", issue.ExistingReportFindings)
	}
	if !strings.Contains(issue.ExistingReportFindings[0].OverviewDetail, "数据库查询耗时升高") || issue.ExistingReportFindings[0].OverviewDetail != issue.ExistingReportFindings[0].Detail {
		t.Fatalf("expected alert_report overview_detail to be included for root cause synthesis, got %#v", issue.ExistingReportFindings[0])
	}
	if !strings.Contains(issue.ExistingReportFindings[0].OverviewReason, "payment-service 调用耗时升高") || issue.ExistingReportFindings[0].OverviewReason != issue.ExistingReportFindings[0].AnalyzeProcess {
		t.Fatalf("expected alert_report overview_reason to be included for eBPF root cause synthesis, got %#v", issue.ExistingReportFindings[0])
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

func TestSystemHealthCleansReasoningFromExistingAlertAnalysis(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
	event := newSystemHealthTestAlert("evt-1", "HighLatency", alertmodel.SeverityCriticalLevel, alertmodel.StatusFiring, now.Add(-20*time.Minute), now)
	event.AlertDirection = `告警发生时分析指向：<details style="color:gray;background-color:#f8f8f8"><summary>Thinking...</summary>**Extracting fault direction** I need a concise answer...`
	event.AnalyzeRunID = "analyze-run-1"
	event.Output = `<details><summary>Thinking...</summary>internal reasoning</details>false`

	inputs := (&handler{}).buildInputsFromRealEvents(nil, 1, []alertmodel.AEventWithWRecord{event}, now.Add(-30*time.Minute), now, telemetryCapabilityStatus{})
	scored := scoreIssues(inputs)
	issues := buildIssues(scored)
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	forbidden := []string{"<details", "Thinking", "Extracting fault", "I need a concise", "告警发生时分析指向"}
	for _, text := range []string{issues[0].RootCause, issues[0].AIConclusion} {
		for _, marker := range forbidden {
			if strings.Contains(text, marker) {
				t.Fatalf("user-facing text should not expose reasoning marker %q: %s", marker, text)
			}
		}
	}
	if len(issues[0].ExistingAlertAnalyses) != 1 {
		t.Fatalf("expected sanitized alert analysis to keep run metadata, got %#v", issues[0].ExistingAlertAnalyses)
	}
	analysisPayload, err := json.Marshal(issues[0].ExistingAlertAnalyses)
	if err != nil {
		t.Fatalf("marshal existing alert analyses: %v", err)
	}
	for _, marker := range forbidden {
		if strings.Contains(string(analysisPayload), marker) {
			t.Fatalf("existing alert analysis should not expose reasoning marker %q: %s", marker, analysisPayload)
		}
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
	aiPayload, err := json.Marshal(buildSystemHealthAIRequest(&resp, scored))
	if err != nil {
		t.Fatalf("marshal ai request: %v", err)
	}
	for _, marker := range forbidden {
		if strings.Contains(string(aiPayload), marker) {
			t.Fatalf("AI request should not expose reasoning marker %q: %s", marker, aiPayload)
		}
	}
}

func TestAlertReportLookupEventIDCandidatesIncludesWorkflowInput(t *testing.T) {
	event := newSystemHealthTestAlert("resolved-event", "HighLatency", alertmodel.SeverityWarnLevel, alertmodel.StatusResolved, time.Now().Add(-time.Minute), time.Now())
	event.Input = "firing-event"

	candidates := alertReportLookupEventIDCandidates(event)

	if len(candidates) != 2 {
		t.Fatalf("expected current event and workflow input candidates, got %#v", candidates)
	}
	if candidates[0].Method != "event_id" || candidates[0].AlertEventID != "resolved-event" {
		t.Fatalf("unexpected first candidate: %#v", candidates[0])
	}
	if candidates[1].Method != "workflow_input" || candidates[1].AlertEventID != "firing-event" {
		t.Fatalf("unexpected second candidate: %#v", candidates[1])
	}

	event.Input = "resolved-event"
	candidates = alertReportLookupEventIDCandidates(event)
	if len(candidates) != 1 {
		t.Fatalf("expected duplicate workflow input to be ignored, got %#v", candidates)
	}
}

func TestBuildAlertHistoryEvidenceResultSummarizesThreeDayStats(t *testing.T) {
	now := time.Date(2026, 6, 17, 10, 30, 0, 0, time.UTC)
	input := reportInput{
		scopeKind: "entry",
		alertHistoryStats: []clickhouse.AlertEventHistoryStats{
			{
				AlertID:              "alert-1",
				RecordCount:          12,
				FiringCount:          6,
				ResolvedCount:        6,
				FirstSeenAt:          now.Add(-48 * time.Hour),
				LastSeenAt:           now.Add(-10 * time.Minute),
				LatestStatus:         alertmodel.StatusResolved,
				MaxDurationSeconds:   300,
				TotalDurationSeconds: 900,
			},
		},
	}

	result := buildAlertHistoryEvidenceResult(input)

	for _, expected := range []string{"过去 3 天", "12 次", "触发记录 6 次", "最长一次持续约 5 分钟", "最近状态为 已恢复"} {
		if !strings.Contains(result, expected) {
			t.Fatalf("expected %q in alert history result: %s", expected, result)
		}
	}
}

func TestBuildIssuesIncludesAlertTrendForThreeDayHeatmap(t *testing.T) {
	now := time.Date(2026, 6, 17, 10, 30, 0, 0, time.UTC)
	event := newSystemHealthTestAlert("evt-1", "HighLatency", alertmodel.SeverityWarnLevel, alertmodel.StatusResolved, now.Add(-20*time.Minute), now.Add(-10*time.Minute))
	input := reportInput{
		key:                   "entry-order-service-post-orders",
		aggregationKey:        "entry::order-service::POST /orders",
		kind:                  issueKindRecoveredAlert,
		title:                 "订单入口延时",
		scope:                 "order-service / POST /orders",
		scopeKind:             "entry",
		startedAt:             formatEventTime(event.CreateTime),
		updatedAt:             formatEventTime(event.UpdateTime),
		events:                []alertmodel.AEventWithWRecord{event},
		availableCalcSources:  []string{"alerts"},
		displaySources:        []string{sourceLabel("alerts")},
		affectedObjects:       affectedObjectsFromEvent(event),
		alertHistoryStartTime: now.Add(-72 * time.Hour),
		alertHistoryEndTime:   now,
		alertHistoryStats: []clickhouse.AlertEventHistoryStats{
			{
				AlertID:              event.AlertID,
				RecordCount:          3,
				FiringCount:          1,
				ResolvedCount:        2,
				FirstSeenAt:          now.Add(-48 * time.Hour),
				LastSeenAt:           now.Add(-10 * time.Minute),
				LatestStatus:         alertmodel.StatusResolved,
				MaxDurationSeconds:   600,
				TotalDurationSeconds: 900,
			},
		},
		alertHistorySegments: []clickhouse.AlertEventRawSegment{
			{
				AlertID:         event.AlertID,
				Name:            "HighLatency",
				Severity:        alertmodel.SeverityWarnLevel,
				Status:          alertmodel.StatusResolved,
				CreateTime:      now.Add(-48 * time.Hour),
				UpdateTime:      now.Add(-48*time.Hour + 5*time.Minute),
				EndTime:         now.Add(-48*time.Hour + 5*time.Minute),
				ReceivedTime:    now.Add(-48*time.Hour + 5*time.Minute),
				DurationSeconds: 300,
			},
			{
				AlertID:         event.AlertID,
				Name:            "HighLatency",
				Severity:        alertmodel.SeverityWarnLevel,
				Status:          alertmodel.StatusResolved,
				CreateTime:      now.Add(-30 * time.Minute),
				UpdateTime:      now.Add(-20 * time.Minute),
				EndTime:         now.Add(-20 * time.Minute),
				ReceivedTime:    now.Add(-20 * time.Minute),
				DurationSeconds: 600,
			},
		},
	}
	input.normalizedEvents = append(input.normalizedEvents, normalizeAlertEvent(event))
	input.alertNames = append(input.alertNames, event.Name)
	input.alertSeverities = append(input.alertSeverities, event.Severity)
	finalizeReportInput(&input, now)

	issues := buildIssues(scoreIssues([]reportInput{input}))
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	trend := issues[0].AlertTrend
	if trend == nil {
		t.Fatalf("expected alertTrend in issue response")
	}
	if trend.Window != "最近 3 天" || trend.StartTime != formatAlertTrendTime(now.Add(-72*time.Hour)) || trend.EndTime != formatAlertTrendTime(now) {
		t.Fatalf("unexpected trend window: %#v", trend)
	}
	if trend.OccurrenceCount != 3 || trend.LongestDurationSeconds != 600 || len(trend.Segments) != 2 {
		t.Fatalf("unexpected trend counters or segments: %#v", trend)
	}
	if len(trend.AlertTypes) == 0 || trend.AlertTypes[0].Name != "请求延时" {
		t.Fatalf("expected latency alert type, got %#v", trend.AlertTypes)
	}
	if trend.Segments[0].AlertID != event.AlertID || trend.Segments[0].DurationSeconds != 300 || trend.Segments[0].Count != 1 {
		t.Fatalf("unexpected first trend segment: %#v", trend.Segments[0])
	}
	payload, err := json.Marshal(issues[0])
	if err != nil {
		t.Fatalf("marshal issue: %v", err)
	}
	if !strings.Contains(string(payload), `"alertTrend"`) || !strings.Contains(string(payload), `"segments"`) {
		t.Fatalf("alertTrend JSON should include segments, got %s", string(payload))
	}
}

func TestBuildAlertTrendSegmentsRebuildsTimelineFromFiringAndResolvedRecords(t *testing.T) {
	now := time.Date(2026, 6, 17, 10, 30, 0, 0, time.UTC)
	alertID := "alert-high-latency"
	windowStart := now.Add(-72 * time.Hour)
	windowEnd := now

	segments := buildAlertTrendSegments(reportInput{
		alertHistorySegments: []clickhouse.AlertEventRawSegment{
			{
				AlertID:      alertID,
				Name:         "HighLatency",
				Severity:     alertmodel.SeverityWarnLevel,
				Status:       alertmodel.StatusFiring,
				CreateTime:   now.Add(-2 * time.Hour),
				UpdateTime:   now.Add(-2 * time.Hour),
				ReceivedTime: now.Add(-2 * time.Hour),
			},
			{
				AlertID:      alertID,
				Name:         "HighLatency",
				Severity:     alertmodel.SeverityWarnLevel,
				Status:       alertmodel.StatusResolved,
				CreateTime:   now.Add(-2 * time.Hour),
				UpdateTime:   now.Add(-90 * time.Minute),
				EndTime:      now.Add(-90 * time.Minute),
				ReceivedTime: now.Add(-90 * time.Minute),
			},
		},
	}, windowStart, windowEnd)

	if len(segments) != 1 {
		t.Fatalf("expected firing/resolved records to become one timeline segment, got %#v", segments)
	}
	segment := segments[0]
	if segment.Status != alertmodel.StatusResolved || segment.IsActive {
		t.Fatalf("expected resolved inactive segment, got %#v", segment)
	}
	if segment.StartTime != formatAlertTrendTime(now.Add(-2*time.Hour)) || segment.EndTime != formatAlertTrendTime(now.Add(-90*time.Minute)) {
		t.Fatalf("unexpected timeline bounds: %#v", segment)
	}
	if segment.DurationSeconds != 1800 || segment.Count != 2 || segment.EventCount != 2 {
		t.Fatalf("unexpected timeline duration/count: %#v", segment)
	}
}

func TestBuildAlertTrendSegmentsMergesOverlappingAlertsForSameRiskPoint(t *testing.T) {
	now := time.Date(2026, 6, 17, 10, 30, 0, 0, time.UTC)
	windowStart := now.Add(-72 * time.Hour)
	windowEnd := now

	segments := buildAlertTrendSegments(reportInput{
		alertHistorySegments: []clickhouse.AlertEventRawSegment{
			{
				AlertID:      "alert-latency",
				Name:         "HighLatency",
				Severity:     alertmodel.SeverityWarnLevel,
				Status:       alertmodel.StatusFiring,
				CreateTime:   now.Add(-2 * time.Hour),
				ReceivedTime: now.Add(-2 * time.Hour),
			},
			{
				AlertID:      "alert-latency",
				Name:         "HighLatency",
				Severity:     alertmodel.SeverityWarnLevel,
				Status:       alertmodel.StatusResolved,
				CreateTime:   now.Add(-2 * time.Hour),
				EndTime:      now.Add(-1 * time.Hour),
				UpdateTime:   now.Add(-1 * time.Hour),
				ReceivedTime: now.Add(-1 * time.Hour),
			},
			{
				AlertID:      "alert-error",
				Name:         "ErrorRateHigh",
				Severity:     alertmodel.SeverityErrorLevel,
				Status:       alertmodel.StatusFiring,
				CreateTime:   now.Add(-90 * time.Minute),
				ReceivedTime: now.Add(-90 * time.Minute),
			},
			{
				AlertID:      "alert-error",
				Name:         "ErrorRateHigh",
				Severity:     alertmodel.SeverityErrorLevel,
				Status:       alertmodel.StatusResolved,
				CreateTime:   now.Add(-90 * time.Minute),
				EndTime:      now.Add(-30 * time.Minute),
				UpdateTime:   now.Add(-30 * time.Minute),
				ReceivedTime: now.Add(-30 * time.Minute),
			},
		},
	}, windowStart, windowEnd)

	if len(segments) != 1 {
		t.Fatalf("expected overlapping alert timelines for one risk point to merge, got %#v", segments)
	}
	segment := segments[0]
	if segment.StartTime != formatAlertTrendTime(now.Add(-2*time.Hour)) || segment.EndTime != formatAlertTrendTime(now.Add(-30*time.Minute)) {
		t.Fatalf("unexpected merged bounds: %#v", segment)
	}
	if segment.DurationSeconds != 5400 || segment.Count != 4 || segment.EventCount != 4 {
		t.Fatalf("unexpected merged duration/count: %#v", segment)
	}
	if segment.Severity != alertmodel.SeverityErrorLevel {
		t.Fatalf("expected merged segment to keep highest severity, got %#v", segment)
	}
	for _, expected := range []string{"alert-latency", "alert-error", "HighLatency", "ErrorRateHigh"} {
		if !strings.Contains(segment.AlertID+" "+segment.AlertName, expected) {
			t.Fatalf("expected merged segment to include %q, got %#v", expected, segment)
		}
	}
}

func TestAlertReportSuggestActionsMergeIntoActionPlan(t *testing.T) {
	actions := parseAlertReportSuggestActions([]string{
		`{"action":"重启异常实例并观察错误率","level":"high"}`,
		`{"action":"重启异常实例并观察错误率","level":"high"}`,
		`{"phase":"排查","target":"下游数据库连接池","action":"补充检查下游数据库连接池","validation":"确认连接池等待下降","risk":"连接池耗尽会继续放大入口延迟","level":"medium"}`,
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
	if len(plan) < 4 {
		t.Fatalf("expected report actions plus system verification steps, got %#v", plan)
	}
	if plan[0].Phase != "原报告高优先级建议" || plan[0].Action != "重启异常实例并观察错误率" {
		t.Fatalf("expected high priority report action first, got %#v", plan[0])
	}
	if plan[0].Source != "alertReport" || plan[1].Source != "alertReport" {
		t.Fatalf("expected report action source to be preserved, got %#v", plan)
	}
	if plan[1].Phase != "排查" || plan[1].Target != "下游数据库连接池" || plan[1].Action != "补充检查下游数据库连接池" {
		t.Fatalf("expected structured report action second, got %#v", plan[1])
	}
	if plan[1].Validation != "确认连接池等待下降。" || plan[1].Risk != "连接池耗尽会继续放大入口延迟" {
		t.Fatalf("expected structured validation and risk from report action, got %#v", plan[1])
	}
	if plan[2].Source != "system" {
		t.Fatalf("expected system verification step to be merged after report actions, got %#v", plan)
	}
}

func TestNormalizeAIActionPlanTargetsReplacesInternalIssueKey(t *testing.T) {
	issue := response.SystemHealthIssue{
		Key:            "recoveredalert-entry-ts-travel2-service-post--api-v1-travel2service-trips-left",
		AggregationKey: "entry::ts-travel2-service::POST /api/v1/travel2service/trips/left",
		Scope:          "ts-travel2-service / POST /api/v1/travel2service/trips/left",
	}
	steps := normalizeAIActionPlanTargets([]response.SystemHealthActionPlanStep{
		{
			Phase:  "排查",
			Target: "recoveredalert-entry-ts-travel2-service-post--api-v1-travel2service-trips-left",
			Action: "检查该入口恢复后的请求量与慢请求样本。",
			Source: "ai",
		},
	}, issue)
	if len(steps) != 1 {
		t.Fatalf("expected one action step, got %#v", steps)
	}
	if steps[0].Target != issue.Scope {
		t.Fatalf("expected internal key target to be replaced by scope, got %#v", steps[0])
	}
}

func TestResolveAIEnrichedActionPlanUsesSynthesizedPlanOverRawReportActions(t *testing.T) {
	issue := response.SystemHealthIssue{
		Key:   "issue-1",
		Scope: "order-service / POST /orders",
		ActionPlan: []response.SystemHealthActionPlanStep{
			{Phase: "原报告建议", Target: "payment-service", Action: "原报告建议：直接扩容 payment-service。", Validation: "确认 payment-service 恢复。", Risk: "原报告风险。", Source: "alertReport"},
			{Phase: "恢复验证", Target: "order-service / POST /orders", Action: "默认回看入口指标。", Validation: "确认无新增告警。", Risk: "默认风险。", Source: "system"},
		},
	}
	reportPlan := []response.SystemHealthActionPlanStep{
		{Phase: "原报告建议", Target: "payment-service", Action: "原报告建议：直接扩容 payment-service。", Validation: "确认 payment-service 恢复。", Risk: "原报告风险。", Source: "alertReport"},
	}
	aiOutput := systemHealthAIIssueOutput{
		Key: "issue-1",
		ActionPlan: []response.SystemHealthActionPlanStep{
			{Phase: "影响确认", Target: "order-service / POST /orders", Action: "结合原报告指向的 payment-service，先确认入口延时和下游调用耗时是否仍同步升高。", Validation: "对比入口延时与 payment-service 调用耗时峰值是否一致。", Risk: "如果下游耗时仍高，入口延时可能继续复发。"},
			{Phase: "恢复验证", Target: "payment-service 与入口 RED 指标", Action: "在处理下游后回看入口 RED 指标和告警状态，确认恢复是否稳定。", Validation: "确认后续窗口没有新的 firing 告警，入口延时回到可接受范围。", Risk: "只看单条告警恢复可能遗漏链路中间节点继续变慢。"},
		},
	}

	plan, mode := resolveAIEnrichedActionPlan(aiOutput, reportPlan, issue.ActionPlan, issue)
	if mode != "ai_synthesized_from_alert_report_actions" {
		t.Fatalf("expected synthesized mode, got %s", mode)
	}
	if len(plan) != 2 {
		t.Fatalf("expected AI synthesized plan only, got %#v", plan)
	}
	for _, step := range plan {
		if step.Source != "ai" {
			t.Fatalf("expected synthesized steps to be sourced from ai, got %#v", plan)
		}
		if strings.HasPrefix(step.Action, "原报告建议：直接扩容") || step.Action == "默认回看入口指标。" {
			t.Fatalf("expected raw report/default actions not to be mechanically appended, got %#v", plan)
		}
	}
}

func TestResolveAIEnrichedActionPlanFallsBackToReportPlanWhenAIHasNoActionPlan(t *testing.T) {
	issue := response.SystemHealthIssue{
		Key:   "issue-1",
		Scope: "order-service / POST /orders",
		ActionPlan: []response.SystemHealthActionPlanStep{
			{Phase: "原报告建议", Target: "payment-service", Action: "检查 payment-service 慢调用。", Validation: "确认慢调用下降。", Risk: "下游变慢会放大入口延迟。", Source: "alertReport"},
		},
	}
	reportPlan := issue.ActionPlan
	plan, mode := resolveAIEnrichedActionPlan(systemHealthAIIssueOutput{Key: "issue-1"}, reportPlan, issue.ActionPlan, issue)
	if mode != "alert_report_structured_fallback" {
		t.Fatalf("expected alert report fallback mode, got %s", mode)
	}
	if len(plan) != 1 || plan[0].Source != "alertReport" || plan[0].Action != "检查 payment-service 慢调用。" {
		t.Fatalf("expected report plan fallback, got %#v", plan)
	}
}

func TestResolveAIEnrichedRootCausePrefersAIOverviewDetailSummary(t *testing.T) {
	item := scoredIssue{
		input: reportInput{
			kind:  issueKindRecoveredAlert,
			scope: "order-service / POST /orders",
			alertReportFindings: []systemHealthAlertReportFinding{
				{
					AlertName:      "HighLatency",
					RootCause:      "报告原始根因：payment-service 响应变慢。",
					AnalyzeProcess: "报告定位到下游调用耗时。",
					Detail:         "alert_report.overview_detail 显示 payment-service 数据库查询耗时升高，order-service 入口等待下游返回导致延时。",
				},
			},
		},
	}
	aiRootCause := "结合方向报告详情，可能根因是 payment-service 数据库查询变慢并放大 order-service 入口等待时间。"

	rootCause, source, reason := resolveAIEnrichedRootCause(aiRootCause, item)
	if !strings.Contains(rootCause, aiRootCause) || !strings.Contains(rootCause, "基于 eBPF 数据") {
		t.Fatalf("expected AI rootCause based on overview_reason to include eBPF context, got %q", rootCause)
	}
	if source != "ai_alert_report_overview_reason" || !strings.Contains(reason, "overview_reason_findings=1") {
		t.Fatalf("expected overview_reason AI source, got source=%s reason=%s", source, reason)
	}
}

func TestResolveAIEnrichedRootCauseUsesOverviewDetailWhenOverviewReasonEmpty(t *testing.T) {
	item := scoredIssue{
		input: reportInput{
			kind:  issueKindRecoveredAlert,
			scope: "order-service / POST /orders",
			alertReportFindings: []systemHealthAlertReportFinding{
				{
					AlertName: "HighLatency",
					Detail:    "alert_report.overview_detail 显示 payment-service 数据库查询耗时升高，order-service 入口等待下游返回导致延时。",
				},
			},
		},
	}
	aiRootCause := "结合方向报告详情，可能根因是 payment-service 数据库查询变慢。"

	rootCause, source, reason := resolveAIEnrichedRootCause(aiRootCause, item)
	if rootCause != aiRootCause {
		t.Fatalf("expected AI rootCause based on overview_detail to be preserved without eBPF prefix, got %q", rootCause)
	}
	if source != "ai_alert_report_overview_detail" || !strings.Contains(reason, "overview_detail_findings=1") {
		t.Fatalf("expected overview_detail AI source, got source=%s reason=%s", source, reason)
	}
}

func TestResolveAIEnrichedRootCauseFallsBackToAlertReportWhenAIEmpty(t *testing.T) {
	item := scoredIssue{
		input: reportInput{
			kind:  issueKindRecoveredAlert,
			scope: "order-service / POST /orders",
			alertReportFindings: []systemHealthAlertReportFinding{
				{
					AlertName:      "HighLatency",
					AnalyzeProcess: "报告定位到下游调用耗时。",
					Detail:         "alert_report.overview_detail 显示 payment-service 数据库查询耗时升高。",
				},
			},
		},
	}

	rootCause, source, reason := resolveAIEnrichedRootCause("", item)
	if !strings.Contains(rootCause, "报告详情") || !strings.Contains(rootCause, "下游调用耗时") {
		t.Fatalf("expected structured report fallback when AI rootCause is empty, got %q", rootCause)
	}
	if source != "alert_report" || !strings.Contains(reason, "ai_root_cause_empty") {
		t.Fatalf("expected alert report fallback source, got source=%s reason=%s", source, reason)
	}
}

func TestResolveAIEnrichedRootCauseSkipsNonDiagnosticAIRootCause(t *testing.T) {
	item := scoredIssue{
		input: reportInput{
			kind:  issueKindRecoveredAlert,
			scope: "order-service / POST /orders",
			alertReportFindings: []systemHealthAlertReportFinding{
				{
					AlertName: "HighLatency",
					Detail:    "方向报告详情显示 payment-service 数据库查询耗时升高。",
				},
			},
		},
	}

	rootCause, source, reason := resolveAIEnrichedRootCause("无法判断根因。", item)
	if strings.Contains(rootCause, "无法判断") || !strings.Contains(rootCause, "payment-service 数据库查询耗时升高") {
		t.Fatalf("expected non-diagnostic AI rootCause to be skipped and report detail fallback used, got %q", rootCause)
	}
	if source != "alert_report" || !strings.Contains(reason, "ai_root_cause_empty") {
		t.Fatalf("expected alert report fallback after non-diagnostic AI rootCause, got source=%s reason=%s", source, reason)
	}
}

func TestRecoveredDefaultActionPlanUsesIssueContext(t *testing.T) {
	item := scoredIssue{
		input: reportInput{
			kind:       issueKindRecoveredAlert,
			title:      "平均请求延时超过1s等 3 类告警",
			scope:      "ts-route-plan-service / POST /api/v1/routeplanservice/routePlan/cheapestRoute",
			alertNames: []string{"平均请求延时超过1s"},
			evidence: []response.SystemHealthEvidence{
				{Type: "告警指标趋势", Result: "查询窗口趋势失败"},
				{Type: "Trace链路样本", Result: "当前窗口匹配 span 样本=20"},
			},
		},
	}

	plan := buildDefaultActionPlan(item)
	if len(plan) < 3 {
		t.Fatalf("expected contextual recovered action plan with multiple steps, got %#v", plan)
	}
	for _, step := range plan {
		if step.Target == "确认有效告警恢复后不再复发。" || step.Action == "查看该影响目标的历史告警和最新指标趋势。" {
			t.Fatalf("recovered fallback should not use the old generic template, got %#v", plan)
		}
		if !strings.Contains(step.Target, "ts-route-plan-service") {
			t.Fatalf("expected action target to include issue scope, got %#v", step)
		}
	}
	if plan[1].Phase != "指标补齐" {
		t.Fatalf("expected unavailable metric evidence to produce metric completion step, got %#v", plan)
	}
	if plan[2].Phase != "链路回看" {
		t.Fatalf("expected trace evidence to produce chain review step, got %#v", plan)
	}
}

func TestRecoveredDefaultActionPlanUsesBaselineStepForCompareAlert(t *testing.T) {
	item := scoredIssue{
		input: reportInput{
			kind:       issueKindRecoveredAlert,
			title:      "平均请求延时相比昨天升高50%以上",
			scope:      "ts-seat-service / POST /api/v1/seatservice/seats/left_tickets",
			alertNames: []string{"平均请求延时相比昨天升高50%以上"},
			evidence: []response.SystemHealthEvidence{
				{Type: "告警指标趋势", Result: "查询窗口趋势失败"},
			},
		},
	}

	plan := buildDefaultActionPlan(item)
	if len(plan) < 2 || plan[1].Phase != "基线核对" {
		t.Fatalf("expected compare-to-yesterday alert to use baseline check, got %#v", plan)
	}
	if !strings.Contains(plan[1].Action, "昨日同窗口") {
		t.Fatalf("expected baseline action to mention yesterday baseline, got %#v", plan[1])
	}
}

func TestParseAlertReportSuggestActionsSupportsChineseKeys(t *testing.T) {
	actions := parseAlertReportSuggestActions([]string{
		`{"阶段":"排查","目标":"下游库存服务","建议":"检查库存服务慢调用和错误日志","验证":"确认入口延时恢复到基线","风险":"库存服务继续变慢会放大入口延时","优先级":"P1"}`,
	}, "HighLatency")
	if len(actions) != 1 {
		t.Fatalf("expected one chinese report action, got %#v", actions)
	}
	action := actions[0]
	if action.Phase != "排查" || action.Target != "下游库存服务" || action.Action != "检查库存服务慢调用和错误日志" {
		t.Fatalf("expected chinese action fields to be parsed, got %#v", action)
	}
	if action.Validation != "确认入口延时恢复到基线" || action.Risk != "库存服务继续变慢会放大入口延时" || action.Level != "p1" {
		t.Fatalf("expected chinese validation/risk/level to be parsed, got %#v", action)
	}
}

func TestAlertReportEvidenceBuildsUserFacingVisibleEvidenceChain(t *testing.T) {
	report := &model.AlertEventReport{
		OverviewReason: "已有分析认为告警有效，延时集中在订单入口。",
		OverviewDetail: "报告详情指向下游库存服务耗时升高。",
		Topology:       `{"nodes":[{"service":"ts-order-service","endpoint":"POST /api/v1/orders"},{"service":"ts-stock-service"}]}`,
		RootCause:      "下游库存服务响应变慢导致入口延时升高。",
		RootCauseOther: []string{
			`{"summary":"库存服务数据库查询排队时间升高，是入口延时放大的补充根因。"}`,
		},
		Data: []string{
			`{"type":"apm","summary":"Trace 慢调用集中在 ts-order-service 到 ts-stock-service 的调用段。"}`,
			`{"type":"metric","summary":"RED 指标显示入口延时在恢复观察窗口内回落。"}`,
			`{"type":"alert","summary":"告警事件多次集中在同一入口，当前已恢复。"}`,
		},
	}
	findings := alertReportFindings(report, "HighLatency")
	if len(findings) != 1 {
		t.Fatalf("expected report evidence to be extracted, got %#v", findings)
	}

	evidence := buildAlertReportEvidence(findings, alertmodel.SeverityCriticalLevel)
	types := make(map[string]string, len(evidence))
	for _, item := range evidence {
		types[item.Type] = item.Result
		if strings.Contains(item.Result, "{") || strings.Contains(item.Result, "raw_tags") || strings.Contains(item.Result, "span_trace") {
			t.Fatalf("expected user-facing report evidence result, got %#v", item)
		}
	}
	if types["方向报告-APM链路"] != "" {
		t.Fatalf("expected APM report evidence to stay hidden from final evidence, got %#v", evidence)
	}
	for _, expected := range []string{"方向报告-指标趋势", "方向报告-告警事件分析"} {
		if types[expected] == "" {
			t.Fatalf("expected %s evidence, got %#v", expected, evidence)
		}
	}
	if types["方向报告-日志分析"] != "" {
		t.Fatalf("expected normal log report evidence to stay hidden, got %#v", evidence)
	}
	if sources := displaySourcesFromEvidence(evidence); containsString(sources, sourceLabel("trace")) {
		t.Fatalf("expected APM report evidence not to add trace display source, got %#v from %#v", sources, evidence)
	}

	aiFindings := buildSystemHealthAIReportFindingInput(findings)
	if len(aiFindings) != 1 || len(aiFindings[0].ReportEvidence) < 2 {
		t.Fatalf("expected report evidence in AI input, got %#v", aiFindings)
	}
	if !strings.Contains(aiFindings[0].OverviewDetail, "下游库存服务耗时升高") || aiFindings[0].OverviewDetail != aiFindings[0].Detail {
		t.Fatalf("expected alert_report overview_detail to be explicit in AI input, got %#v", aiFindings[0])
	}
	if !strings.Contains(aiFindings[0].OverviewReason, "延时集中在订单入口") || aiFindings[0].OverviewReason != aiFindings[0].AnalyzeProcess {
		t.Fatalf("expected alert_report overview_reason to be explicit in AI input, got %#v", aiFindings[0])
	}
	if len(aiFindings[0].RootCauseOther) != 1 || !strings.Contains(aiFindings[0].RootCauseOther[0], "库存服务数据库查询排队时间升高") {
		t.Fatalf("expected alert_report root_cause_other to be explicit in AI input, got %#v", aiFindings[0])
	}
}

func TestAlertReportEvidenceSkipsLogAnalysisWithoutValidLogData(t *testing.T) {
	report := &model.AlertEventReport{
		OverviewReason: "已有分析认为告警有效，延时集中在订单入口。",
		OverviewDetail: "报告详情指向下游库存服务耗时升高。",
		Data: []string{
			`{"type":"log","summary":"无有效日志数据，无法进行日志分析。"}`,
			`{"type":"metric","summary":"RED 指标显示入口延时在恢复观察窗口内回落。"}`,
			`{"type":"alert","summary":"告警事件多次集中在同一入口，当前已恢复。"}`,
		},
	}

	findings := alertReportFindings(report, "HighLatency")
	if len(findings) != 1 {
		t.Fatalf("expected report finding to be extracted, got %#v", findings)
	}
	evidence := buildAlertReportEvidence(findings, alertmodel.SeverityWarnLevel)
	for _, item := range evidence {
		if item.Type == "方向报告-日志分析" {
			t.Fatalf("expected log report evidence without valid log data to stay hidden, got %#v", evidence)
		}
	}
	if sources := displaySourcesFromEvidence(evidence); containsString(sources, sourceLabel("logs")) {
		t.Fatalf("expected invalid log report evidence not to add log display source, got %#v from %#v", sources, evidence)
	}
}

func TestAlertReportEvidenceKeepsOnlyExplicitErrorLogSignals(t *testing.T) {
	report := &model.AlertEventReport{
		OverviewReason: "已有分析认为告警有效，延时集中在订单入口。",
		OverviewDetail: "报告详情指向下游库存服务耗时升高。",
		Data: []string{
			`{"type":"log","summary":"库存服务日志出现 ERROR timeout，车次 G1234、订单 orderId=1234567890123 返回 5xx，错误堆栈指向数据库查询超时。"}`,
			`{"type":"metric","summary":"RED 指标显示入口延时升高。"}`,
		},
	}

	findings := alertReportFindings(report, "HighLatency")
	if len(findings) != 1 {
		t.Fatalf("expected report finding to be extracted, got %#v", findings)
	}
	evidence := buildAlertReportEvidence(findings, alertmodel.SeverityErrorLevel)
	var logEvidence response.SystemHealthEvidence
	for _, item := range evidence {
		if item.Type == "方向报告-日志分析" {
			logEvidence = item
			break
		}
	}
	if logEvidence.Result == "" || !strings.Contains(logEvidence.Result, "ERROR timeout") {
		t.Fatalf("expected explicit error log evidence to be kept, got %#v", evidence)
	}
	if strings.Contains(logEvidence.Result, "G1234") || strings.Contains(logEvidence.Result, "1234567890123") || strings.Contains(logEvidence.Result, "orderId=1234567890123") {
		t.Fatalf("expected business log content to be redacted, got %#v", logEvidence)
	}
	if !strings.Contains(logEvidence.Result, "车次已脱敏") || !strings.Contains(logEvidence.Result, "orderId=已脱敏") {
		t.Fatalf("expected redacted log evidence to preserve diagnostic context, got %#v", logEvidence)
	}
	if sources := displaySourcesFromEvidence(evidence); !containsString(sources, sourceLabel("logs")) {
		t.Fatalf("expected explicit error log evidence to add log display source, got %#v from %#v", sources, evidence)
	}
}

func TestAlertReportFindingsFiltersNonDiagnosticOverviewReason(t *testing.T) {
	report := &model.AlertEventReport{
		OverviewReason: "无明确故障方向。",
	}

	findings := alertReportFindings(report, "HighLatency")
	if len(findings) != 0 {
		t.Fatalf("expected non-diagnostic overview_reason to be filtered, got %#v", findings)
	}

	report.OverviewDetail = "方向报告详情显示 payment-service 数据库查询耗时升高。"
	findings = alertReportFindings(report, "HighLatency")
	if len(findings) != 1 {
		t.Fatalf("expected valid overview_detail to keep finding, got %#v", findings)
	}
	if findings[0].AnalyzeProcess != "" || !strings.Contains(findings[0].Detail, "payment-service 数据库查询耗时升高") {
		t.Fatalf("expected non-diagnostic overview_reason to be removed while keeping detail, got %#v", findings[0])
	}
}

func TestBuildRootCauseFromAlertReportsSkipsNonDiagnosticAnalyzeProcess(t *testing.T) {
	rootCause := buildRootCauseFromAlertReports([]systemHealthAlertReportFinding{
		{AlertName: "HighLatency", AnalyzeProcess: "未发现明显故障方向。"},
	})
	if rootCause != "" {
		t.Fatalf("expected non-diagnostic analyzeProcess to be skipped, got %q", rootCause)
	}
}

func TestAlertReportTopologyRoleDetectsEntryEndpoint(t *testing.T) {
	now := time.Now()
	event := newSystemHealthTestAlert("evt-entry", "HighLatency", alertmodel.SeverityWarnLevel, alertmodel.StatusResolved, now.Add(-10*time.Minute), now)
	report := &model.AlertEventReport{
		Tags: map[string]string{"alertEventId": "evt-entry"},
		Topology: `{
			"current":{"service":"order-service","endpoint":"POST /orders"},
			"parents":[],
			"children":[{"service":"stock-service"}]
		}`,
	}

	role := alertReportTopologyRole(report, event)
	if role.Role != "entry" || role.ParentCount != 0 || role.ChildCount != 1 {
		t.Fatalf("expected entry topology role, got %#v", role)
	}
	if !strings.Contains(role.Summary, "未发现上游调用方") || !strings.Contains(role.Summary, "优先核对该请求") {
		t.Fatalf("expected user-facing entry topology summary, got %s", role.Summary)
	}
}

func TestAlertReportTopologyRoleDetectsInternalServiceEndpoint(t *testing.T) {
	now := time.Now()
	event := newSystemHealthTestAlert("evt-internal", "HighLatency", alertmodel.SeverityWarnLevel, alertmodel.StatusResolved, now.Add(-10*time.Minute), now)
	report := &model.AlertEventReport{
		Tags: map[string]string{"alertEventId": "evt-internal"},
		Topology: `{
			"current":{"service":"order-service","endpoint":"POST /orders"},
			"parents":[{"service":"gateway-service","endpoint":"GET /checkout"}],
			"children":[{"service":"stock-service"}]
		}`,
	}

	role := alertReportTopologyRole(report, event)
	if role.Role != "internalEndpoint" || role.ParentCount != 1 {
		t.Fatalf("expected internal endpoint topology role, got %#v", role)
	}
	input := reportInput{
		key:                      "entry-order-service-post-orders",
		aggregationKey:           "entry::order-service::POST /orders",
		kind:                     issueKindRecoveredAlert,
		scope:                    "order-service / POST /orders",
		scopeKind:                "entry",
		objectCriticality:        90,
		alertReportTopologyRoles: []systemHealthAlertReportTopologyRole{role},
	}
	input.mergeReasons = buildMergeReasons(input)
	applyAlertReportTopologyRoleToInput(nil, &input)
	if input.scopeKind != "serviceEndpoint" {
		t.Fatalf("expected topology to correct entry scope to serviceEndpoint, got %s", input.scopeKind)
	}
	if !strings.Contains(strings.Join(input.mergeReasons, "、"), "上游调用方") {
		t.Fatalf("expected merge reasons to mention topology upstream, got %#v", input.mergeReasons)
	}
	result := buildAlertEventEvidenceResult(input)
	if strings.Contains(result, "业务入口") || strings.Contains(result, "入口侧接口") {
		t.Fatalf("expected internal endpoint evidence not to call it an entry, got %s", result)
	}
	if !strings.Contains(result, "服务接口") {
		t.Fatalf("expected internal endpoint evidence to use service endpoint wording, got %s", result)
	}
}

func TestAlertReportFindingsCleanReasoningBlocks(t *testing.T) {
	report := &model.AlertEventReport{
		OverviewReason: `<details><summary>Thinking...</summary>internal reasoning</details>已有分析认为告警有效，延时集中在订单入口。`,
		OverviewDetail: `报告详情指向下游库存服务耗时升高。<details><summary>Thinking...</summary>debug thought</details>`,
		RootCause:      `<details><summary>Thinking...</summary>**Extracting fault direction** I need a concise answer</details>下游库存服务响应变慢导致入口延时升高。`,
		Data: []string{
			`{"type":"apm","summary":"<details><summary>Thinking...</summary>internal</details>Trace 慢调用集中在 ts-order-service 到 ts-stock-service 的调用段。"}`,
			`{"type":"metric","summary":"RED 指标显示入口延时在恢复观察窗口内回落。"}`,
		},
	}
	findings := alertReportFindings(report, "HighLatency")
	if len(findings) != 1 {
		t.Fatalf("expected one cleaned finding, got %#v", findings)
	}
	if !strings.Contains(findings[0].RootCause, "下游库存服务响应变慢") {
		t.Fatalf("expected cleaned root cause to keep useful report conclusion, got %#v", findings[0])
	}
	evidence := buildAlertReportEvidence(findings, alertmodel.SeverityCriticalLevel)
	payload, err := json.Marshal(struct {
		Findings []systemHealthAlertReportFinding `json:"findings"`
		Evidence []response.SystemHealthEvidence  `json:"evidence"`
	}{
		Findings: findings,
		Evidence: evidence,
	})
	if err != nil {
		t.Fatalf("marshal report evidence: %v", err)
	}
	for _, marker := range []string{"<details", "Thinking", "Extracting fault", "I need a concise", "internal reasoning"} {
		if strings.Contains(string(payload), marker) {
			t.Fatalf("report finding should not expose reasoning marker %q: %s", marker, payload)
		}
	}
}

func TestCleanUserFacingReportTextStripsEnglishReasoningPrefix(t *testing.T) {
	text := cleanUserFacingReportText("Summarizing log status I need to provide an answer in Chinese. 当前日志以 INFO 为主，无明显报错。")

	if strings.Contains(text, "Summarizing") || strings.Contains(text, "I need") {
		t.Fatalf("expected internal English reasoning prefix to be stripped, got %q", text)
	}
	if !strings.Contains(text, "当前日志以 INFO 为主") {
		t.Fatalf("expected user-facing Chinese conclusion to be kept, got %q", text)
	}
}

func TestMergeActionPlanStepsNormalizesValidationAsCheckItem(t *testing.T) {
	plan := mergeActionPlanSteps([]response.SystemHealthActionPlanStep{
		{
			Phase:      "验证",
			Target:     "stations/idlist 接口",
			Action:     "查看站点服务日志和慢请求样本。",
			Validation: "日志无持续慢请求或异常堆积。",
			Risk:       "当前缺少调用关系。",
		},
	}, nil)
	if len(plan) != 1 {
		t.Fatalf("expected one normalized action step, got %#v", plan)
	}
	expected := "检查日志中是否仍存在持续慢请求或异常堆积，并确认恢复后不再新增。"
	if plan[0].Validation != expected {
		t.Fatalf("expected validation to be phrased as a check item, got %q", plan[0].Validation)
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
		for _, expected := range []string{
			"任务：把当前窗口的系统健康状态",
			"普通用户可理解、可处置的系统健康报告",
			"各字段职责必须清晰",
			"summary 用于历史列表或空状态",
			"避免在 analysis、metricAnalysis、rootCause 和 actionPlan 中重复同一段结论",
			"issue.title",
			"真实服务名",
			"alertNames",
			"不要根据服务名推断业务功能",
			"issue.title 是风险名称，不是分析结论",
			"issue.title 优先使用“{对象} 存在{问题类型}问题”",
			"状态前缀“已恢复：{对象} 存在{问题类型}问题”",
			"issue.title 不要包含“恢复观察”",
			"不要返回“xxx 等 N 类告警”",
			"issue.analysis 只说明当前风险是什么",
			"不复述 metricAnalysis 的指标细节",
			"不复述 rootCause 的完整根因判断",
			"issue.metricAnalysis",
			"指标名",
			"metricName",
			"ruleName",
			"不是指标名",
			"服务/请求路径",
			"时间段",
			"指标表现",
			"当前风险含义",
			"不要输出 PQL",
			"PromQL",
			"rate(",
			"sum by(",
			"query 表达式",
			"issue.rootCause 来源优先级",
			"existingReportFindings",
			"基于 eBPF 数据",
			"无明确故障方向",
			"现有证据暂不能归因",
			"不要从正常业务日志中推断问题",
			"请求高频、批量查询、返回成功、业务字段取值、默认值",
			"不能输出“从日志中可以看出”",
			"日志分析涉及业务内容时必须脱敏",
			"不要输出车次、订单号、用户号、手机号、证件号",
			"请求北极星表示单次请求平均耗时分布",
			"P4 invalidAlert",
			"actionPlan 不是从 0 开始排查",
			"沿着已有分析继续定位尚未闭合的问题",
			"基于已有根因/指标/日志/链路结论执行缓解",
			"不要输出泛化起手式排查",
			"existingReportActions 是参考来源",
			"不能原样复制或机械追加",
			"target 不得使用 issue.key",
			"validation 必须使用“确认/检查/观察/回看/对比”",
		} {
			if !strings.Contains(gotRequest.Instructions, expected) {
				t.Fatalf("expected metricAnalysis instruction to mention %q, got: %s", expected, gotRequest.Instructions)
			}
		}
		if !strings.Contains(gotRequest.Input, "payment-service 下游变慢") {
			t.Fatalf("expected existing alert analysis in OpenAI input: %s", gotRequest.Input)
		}
		var gotInput systemHealthAISummaryRequest
		if err := json.Unmarshal([]byte(gotRequest.Input), &gotInput); err != nil {
			t.Fatalf("decode AI input payload: %v", err)
		}
		if gotInput.Window != "15m" || len(gotInput.Issues) != 1 {
			t.Fatalf("unexpected AI input shape: %#v", gotInput)
		}
		gotIssue := gotInput.Issues[0]
		if gotIssue.Scope != "order-service / POST /orders" || gotIssue.StartedAt != "2026-06-16 10:30" {
			t.Fatalf("expected user-facing metric context to include service/path and time period, got %#v", gotIssue)
		}
		if len(gotIssue.AlertNames) != 1 || gotIssue.AlertNames[0] != "HighLatency" {
			t.Fatalf("expected real alert names in AI input, got %#v", gotIssue.AlertNames)
		}
		if len(gotIssue.Evidence) != 1 || gotIssue.Evidence[0].Metric == nil {
			t.Fatalf("expected metric trend evidence in AI input, got %#v", gotIssue.Evidence)
		}
		if len(gotIssue.ExistingReportActions) != 1 ||
			gotIssue.ExistingReportActions[0].Target != "payment-service" ||
			!strings.Contains(gotIssue.ExistingReportActions[0].Action, "扩容 payment-service") ||
			!strings.Contains(gotIssue.ExistingReportActions[0].Validation, "入口延迟") {
			t.Fatalf("expected existing report actions to be included for synthesis, got %#v", gotIssue.ExistingReportActions)
		}
		if len(gotIssue.ExistingReportFindings) != 1 ||
			!strings.Contains(gotIssue.ExistingReportFindings[0].OverviewDetail, "数据库查询耗时升高") ||
			gotIssue.ExistingReportFindings[0].OverviewDetail != gotIssue.ExistingReportFindings[0].Detail {
			t.Fatalf("expected alert_report overview_detail to be included for root cause synthesis, got %#v", gotIssue.ExistingReportFindings)
		}
		if !strings.Contains(gotIssue.ExistingReportFindings[0].OverviewReason, "payment-service 调用耗时升高") ||
			gotIssue.ExistingReportFindings[0].OverviewReason != gotIssue.ExistingReportFindings[0].AnalyzeProcess {
			t.Fatalf("expected alert_report overview_reason to be included for eBPF root cause synthesis, got %#v", gotIssue.ExistingReportFindings)
		}
		gotMetricEvidence := gotIssue.Evidence[0]
		gotMetric := gotMetricEvidence.Metric
		if !strings.Contains(gotMetricEvidence.Result, "15分钟窗口") ||
			!strings.Contains(gotMetricEvidence.Result, "order-service / POST /orders") ||
			gotMetric.LastValue <= gotMetric.FirstValue ||
			gotMetric.ChangePercent == nil ||
			*gotMetric.ChangePercent <= 0 ||
			gotMetric.MetricName != "平均请求延时" ||
			gotMetric.RuleName != "HighLatency" ||
			!strings.Contains(gotMetric.Message, "当前仍在触发") {
			t.Fatalf("expected metric evidence to carry performance and current meaning, got evidence=%#v metric=%#v", gotMetricEvidence, gotMetric)
		}
		assertNoUserFacingMetricDebugContent(t, gotMetricEvidence.Result)
		assertNoUserFacingMetricDebugContent(t, gotMetric.Message)
		assertNoUserFacingMetricDebugContent(t, gotRequest.Input)
		aiOutput, err := json.Marshal(systemHealthAISummaryResponse{
			Summary: "当前 order-service 支付入口存在高优先级延迟问题，应先控制入口影响。",
			Issues: []systemHealthAIIssueOutput{
				{
					Key:            "issue-1",
					Title:          "order-service 入口延迟风险",
					Analysis:       "order-service / POST /orders 延迟持续升高，已有告警分析指向 payment-service。",
					RootCause:      "payment-service 下游响应变慢导致 order-service 入口延迟。",
					MetricAnalysis: "入口延迟指标仍在升高，说明当前告警仍有指标支撑。",
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

	changePercent := 100.0
	payload := systemHealthAISummaryRequest{
		Window: "15m",
		Report: systemHealthAIReportInput{
			Status:  "warning",
			Score:   68,
			Summary: "结构化摘要",
		},
		Issues: []systemHealthAIIssueInput{
			{
				Key:        "issue-1",
				Priority:   "P1",
				Title:      "订单入口延迟升高",
				AlertNames: []string{"HighLatency"},
				Status:     "active",
				Scope:      "order-service / POST /orders",
				Impact:     "用户下单入口响应变慢。",
				StartedAt:  "2026-06-16 10:30",
				Evidence: []systemHealthAIEvidenceInput{
					{
						Type:   "告警指标趋势",
						Source: "Prometheus",
						Result: "15分钟窗口内 order-service / POST /orders 入口延迟从 0.8s 升至 1.6s，当前仍在触发，说明用户请求体验仍受影响。",
						Metric: &systemHealthAIMetricInput{
							MetricName:     "平均请求延时",
							RuleName:       "HighLatency",
							SeriesMatched:  1,
							SeriesReturned: 1,
							PointCount:     6,
							FirstValue:     0.8,
							LastValue:      1.6,
							MinValue:       0.8,
							MaxValue:       1.6,
							Delta:          0.8,
							ChangePercent:  &changePercent,
							EventScoped:    true,
							Message:        "指标已匹配 order-service / POST /orders，当前仍在触发，风险仍有指标支撑。",
						},
					},
				},
				ExistingAlertAnalyses: []systemHealthAIAlertAnalysisInput{
					{AlertDirection: "payment-service 下游变慢", AnalyzeRunID: "analyze-run-1"},
				},
				ExistingReportFindings: []systemHealthAIReportFindingInput{
					{
						AlertName:      "HighLatency",
						RootCause:      "payment-service 响应变慢导致 order-service 入口延迟。",
						AnalyzeProcess: "方向报告定位到 payment-service 调用耗时升高。",
						OverviewReason: "方向报告定位到 payment-service 调用耗时升高。",
						Detail:         "alert_report.overview_detail 显示 payment-service 数据库查询耗时升高，入口等待下游返回导致订单接口延迟。",
						OverviewDetail: "alert_report.overview_detail 显示 payment-service 数据库查询耗时升高，入口等待下游返回导致订单接口延迟。",
					},
				},
				ExistingReportActions: []systemHealthAIReportActionInput{
					{
						AlertName:  "HighLatency",
						Level:      "high",
						Phase:      "原报告止血",
						Target:     "payment-service",
						Action:     "扩容 payment-service 实例，并观察入口延迟是否恢复。",
						Validation: "确认扩容后入口延迟下降。",
						Risk:       "下游容量不足会继续放大入口延迟。",
					},
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
	if aiResp.Issues[0].Title != "order-service 入口延迟风险" {
		t.Fatalf("expected title from AI response, got %#v", aiResp.Issues[0])
	}
	if aiResp.Issues[0].MetricAnalysis != "入口延迟指标仍在升高，说明当前告警仍有指标支撑。" {
		t.Fatalf("expected metricAnalysis from AI response, got %#v", aiResp.Issues[0])
	}
	assertNoUserFacingMetricDebugContent(t, aiResp.Issues[0].MetricAnalysis)
}

func TestCallSystemHealthAIEnrichmentSplitsIssueRequestsInParallel(t *testing.T) {
	var inFlight int32
	var maxInFlight int32
	var summaryCalls int32
	var issueCalls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&inFlight, 1)
		defer atomic.AddInt32(&inFlight, -1)
		for {
			observed := atomic.LoadInt32(&maxInFlight)
			if current <= observed || atomic.CompareAndSwapInt32(&maxInFlight, observed, current) {
				break
			}
		}
		time.Sleep(80 * time.Millisecond)

		var gotRequest openAIResponsesRequest
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		var gotInput systemHealthAISummaryRequest
		if err := json.Unmarshal([]byte(gotRequest.Input), &gotInput); err != nil {
			t.Fatalf("decode AI input payload: %v", err)
		}
		aiOutput := systemHealthAISummaryResponse{
			Summary: "当前系统处于警告状态，多个风险需要结合指标趋势和告警分析继续确认。",
		}
		if len(gotInput.Issues) == 0 {
			atomic.AddInt32(&summaryCalls, 1)
		} else if len(gotInput.Issues) == 1 {
			atomic.AddInt32(&issueCalls, 1)
			issue := gotInput.Issues[0]
			aiOutput.Issues = []systemHealthAIIssueOutput{
				{
					Key:            issue.Key,
					Title:          issue.Title,
					Analysis:       issue.Scope + " 当前风险已完成智能分析。",
					RootCause:      "现有证据暂不能归因，需要结合告警指标或明确错误日志继续确认。",
					MetricAnalysis: "",
					ActionPlan: []response.SystemHealthActionPlanStep{
						{
							Phase:      "确认影响",
							Target:     issue.Scope,
							Action:     "回看该风险对象的告警指标和请求表现，确认是否仍存在用户影响。",
							Validation: "确认后续窗口未继续出现同类告警。",
							Risk:       "若同类告警复发，说明风险仍需进一步排查。",
						},
						{
							Phase:      "复发观察",
							Target:     issue.Scope,
							Action:     "对比恢复后的告警指标趋势和日志错误信号，确认是否存在反复波动。",
							Validation: "观察后续窗口没有新的异常峰值。",
							Risk:       "只看单次恢复可能遗漏短时波动。",
						},
					},
				},
			}
		} else {
			t.Fatalf("expected summary-only or single-issue AI input, got %d issues", len(gotInput.Issues))
		}
		outputBytes, err := json.Marshal(aiOutput)
		if err != nil {
			t.Fatalf("marshal AI output: %v", err)
		}
		_ = json.NewEncoder(w).Encode(openAIResponsesResponse{
			ID:         "resp-" + systemHealthShortHash(gotRequest.Input),
			OutputText: string(outputBytes),
		})
	}))
	defer server.Close()

	payload := systemHealthAISummaryRequest{
		Window: "30m",
		Report: systemHealthAIReportInput{
			Status:  "warning",
			Score:   70,
			Summary: "结构化摘要",
		},
		Issues: []systemHealthAIIssueInput{
			{Key: "issue-1", Priority: "P3", Title: "service-a 存在请求延时问题", Scope: "service-a", Status: "recovered"},
			{Key: "issue-2", Priority: "P3", Title: "service-b 存在请求延时问题", Scope: "service-b", Status: "recovered"},
			{Key: "issue-3", Priority: "P3", Title: "service-c 存在请求延时问题", Scope: "service-c", Status: "recovered"},
			{Key: "issue-4", Priority: "P3", Title: "service-d 存在请求延时问题", Scope: "service-d", Status: "recovered"},
		},
	}

	aiResp, stats, err := callSystemHealthAIEnrichment(nil, server.URL, "test-key", "gpt-5.5", 3, payload)
	if err != nil {
		t.Fatalf("call parallel AI enrichment: %v, stats=%#v", err, stats)
	}
	if stats.Mode != "parallel_per_issue" || stats.TotalCalls != 5 || stats.SucceededCalls != 5 || stats.SucceededIssues != 4 {
		t.Fatalf("unexpected AI stats: %#v", stats)
	}
	if atomic.LoadInt32(&summaryCalls) != 1 || atomic.LoadInt32(&issueCalls) != 4 {
		t.Fatalf("expected one summary call and four issue calls, got summary=%d issue=%d", summaryCalls, issueCalls)
	}
	if atomic.LoadInt32(&maxInFlight) < 2 {
		t.Fatalf("expected concurrent AI requests, max in-flight=%d", maxInFlight)
	}
	if len(aiResp.Issues) != 4 || aiResp.Issues[0].Key != "issue-1" || aiResp.Issues[3].Key != "issue-4" {
		t.Fatalf("expected issue outputs to preserve request order, got %#v", aiResp.Issues)
	}
	if status := systemHealthAIAnalysisStatus(stats, len(payload.Issues)); status != "succeeded" {
		t.Fatalf("expected succeeded AI status, got %s", status)
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
	sampleStart := time.Date(2026, 6, 16, 10, 30, 0, 0, time.Local)
	results := []alertMetricSeries{
		{
			Labels: map[string]string{
				"serviceName": "order-service",
				"endpoint":    "POST /orders",
				"namespace":   "prod",
			},
			Values: []prometheus.Points{
				{TimeStamp: sampleStart.UnixNano() / 1e3, Value: 1},
				{TimeStamp: sampleStart.Add(5*time.Minute).UnixNano() / 1e3, Value: 2},
			},
		},
		{
			Labels: map[string]string{
				"serviceName": "catalog",
				"endpoint":    "GET /items",
			},
			Values: []prometheus.Points{
				{TimeStamp: sampleStart.UnixNano() / 1e3, Value: 10},
				{TimeStamp: sampleStart.Add(5*time.Minute).UnixNano() / 1e3, Value: 20},
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
	if metric.Series[0].Points[0].Timestamp != sampleStart.UnixNano()/1e3 || metric.Series[0].Points[0].Value != 1 {
		t.Fatalf("unexpected first chart point: %#v", metric.Series[0].Points[0])
	}

	analysis := buildMetricAnalysisForIssueContext("order-service / POST /orders", "active", 1, []response.SystemHealthEvidence{{Type: "告警指标趋势", Metric: metric}})
	for _, expected := range []string{"order-service / POST /orders", "请求延时", "2026-06-16 10:30", "2026-06-16 10:35", "升高", "当前风险仍有指标支撑"} {
		if !strings.Contains(analysis, expected) {
			t.Fatalf("expected readable metric analysis to contain %q, got %q", expected, analysis)
		}
	}
	if strings.Contains(analysis, "HighLatency") {
		t.Fatalf("metric analysis should not use alertname as metric name, got %q", analysis)
	}
	assertNoUserFacingMetricDebugContent(t, analysis)

	fallbackAnalysis := buildMetricAnalysis([]response.SystemHealthEvidence{{Type: "告警指标趋势", Metric: metric}})
	if !strings.Contains(fallbackAnalysis, "该影响目标的请求延时") || strings.Contains(fallbackAnalysis, "HighLatency") {
		t.Fatalf("expected fallback metric analysis to keep user-facing subject, got %q", fallbackAnalysis)
	}
	assertNoUserFacingMetricDebugContent(t, fallbackAnalysis)

	aiRequest := buildSystemHealthAIRequest(&response.SystemHealthReportResp{
		Report: response.SystemHealthReport{Window: "15m", Status: "warning"},
		Issues: []response.SystemHealthIssue{
			{
				Key:          "issue-1",
				Scope:        "order-service / POST /orders",
				Status:       "active",
				ActiveAlerts: 1,
				Evidence: []response.SystemHealthEvidence{{
					Type:   "告警指标趋势",
					Result: "order-service / POST /orders 延时指标在窗口内升高。",
					Metric: metric,
				}},
			},
		},
	}, nil)
	if len(aiRequest.Issues) != 1 {
		t.Fatalf("expected one AI request issue, got %#v", aiRequest.Issues)
	}
	aiIssue := aiRequest.Issues[0]
	for _, expected := range []string{"order-service / POST /orders", "请求延时", "2026-06-16 10:30", "当前风险仍有指标支撑"} {
		if !strings.Contains(aiIssue.MetricNarrativeContext, expected) {
			t.Fatalf("expected AI metric narrative context to contain %q, got %q", expected, aiIssue.MetricNarrativeContext)
		}
	}
	if strings.Contains(aiIssue.MetricNarrativeContext, "HighLatency") {
		t.Fatalf("AI metric narrative should not use alertname as metric name, got %q", aiIssue.MetricNarrativeContext)
	}
	if len(aiIssue.Evidence) != 1 || aiIssue.Evidence[0].Metric == nil {
		t.Fatalf("expected AI metric evidence, got %#v", aiIssue.Evidence)
	}
	aiMetric := aiIssue.Evidence[0].Metric
	if aiMetric.MetricName != "请求延时" || aiMetric.RuleName != "HighLatency" || aiMetric.Trend != "整体升高" || aiMetric.StartAt != "2026-06-16 10:30" || aiMetric.EndAt != "2026-06-16 10:35" || aiMetric.PeakAt != "2026-06-16 10:35" {
		t.Fatalf("expected AI metric time context, got %#v", aiMetric)
	}
	assertNoUserFacingMetricDebugContent(t, aiIssue.MetricNarrativeContext)
}

func TestInferMetricNameFromAlertRuleName(t *testing.T) {
	cases := []struct {
		ruleName   string
		expression string
		want       string
	}{
		{ruleName: "平均请求延时超过1s", want: "平均请求延时"},
		{ruleName: "平均请求延时相比昨天升高50%以上", want: "请求延时日同比"},
		{ruleName: "请求错误率超过0%", want: "请求错误率"},
		{ruleName: "请求错误率相比昨天升高20%以上", want: "请求错误率日同比"},
		{ruleName: "日志错误数相比昨天升高20%以上", want: "日志错误数日同比"},
		{ruleName: "HighLatency", expression: "sum(rate(http_request_duration_seconds_sum[5m]))", want: "请求延时"},
		{ruleName: "UnknownBusinessRule", want: "相关指标"},
	}
	for _, tc := range cases {
		if got := inferMetricNameFromRule(tc.ruleName, tc.expression); got != tc.want {
			t.Fatalf("inferMetricNameFromRule(%q, %q)=%q, want %q", tc.ruleName, tc.expression, got, tc.want)
		}
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
		kind:      issueKindRecoveredAlert,
		scope:     "order-service / POST /orders",
		scopeKind: "entry",
		events:    []alertmodel.AEventWithWRecord{first, second, third},
	}
	for _, event := range input.events {
		normalized := normalizeAlertEvent(event)
		input.normalizedEvents = append(input.normalizedEvents, normalized)
		input.alertNames = append(input.alertNames, event.Name)
		input.alertSeverities = append(input.alertSeverities, normalized.Severity)
	}
	finalizeReportInput(&input, now)

	result := buildAlertEventEvidenceResult(input)
	for _, expected := range []string{"当前窗口多次出现", "当前已恢复", "共同指向同一入口侧接口", "不是单条告警孤立触发"} {
		if !strings.Contains(result, expected) {
			t.Fatalf("expected %q in evidence result: %s", expected, result)
		}
	}
	for _, unexpected := range []string{"事件规律=", "状态=", "聚合对象=", "去重告警实例", "去重实例", "同一对象聚合"} {
		if strings.Contains(result, unexpected) {
			t.Fatalf("unexpected raw alert field %q in evidence result: %s", unexpected, result)
		}
	}
}

func TestIsLatencyIssueInputDetectsLatencySignals(t *testing.T) {
	cases := []struct {
		name  string
		input reportInput
		want  bool
	}{
		{
			name:  "chinese latency title",
			input: reportInput{title: "平均请求延时相比昨天升高50%以上"},
			want:  true,
		},
		{
			name: "english latency alert name",
			input: reportInput{
				title:      "业务告警",
				alertNames: []string{"HighLatency"},
			},
			want: true,
		},
		{
			name:  "timeout impact",
			input: reportInput{impact: "users see timeout on checkout requests"},
			want:  true,
		},
		{
			name: "non latency issue",
			input: reportInput{
				title:      "CPU usage high",
				scope:      "node-a",
				impact:     "worker node resource pressure",
				alertNames: []string{"ErrorRateHigh"},
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isLatencyIssueInput(tc.input); got != tc.want {
				t.Fatalf("isLatencyIssueInput()=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestRequestPolarisSelectorEscapesMatchersAndDefaultsEndpoint(t *testing.T) {
	got := requestPolarisSelector("order.service(blue)", "POST /api/v1/orders?x=1", "cluster-a")
	want := `svc_name=~"order\\.service\\(blue\\)",content_key=~"POST /api/v1/orders\\?x=1",cluster_id=~"cluster-a"`
	if got != want {
		t.Fatalf("unexpected selector:\ngot  %s\nwant %s", got, want)
	}

	got = requestPolarisSelector("order-service", "", "")
	want = `svc_name=~"order-service",content_key=~".*"`
	if got != want {
		t.Fatalf("expected empty endpoint to use wildcard and omit cluster matcher, got %s", got)
	}
}

func TestRequestPolarisMetricSpecsIncludeRequestResourceDimensions(t *testing.T) {
	specs := requestPolarisMetricSpecs("order-service", "POST /orders", "cluster-a", "5m", time.Minute)
	byKey := make(map[string]requestPolarisMetricTrend, len(specs))
	for _, spec := range specs {
		byKey[spec.Key] = spec
	}

	for _, key := range []string{"net", "epoll", "cpu", "latency", "tps"} {
		if _, ok := byKey[key]; !ok {
			t.Fatalf("expected request Polaris spec %q in %#v", key, specs)
		}
	}
	if byKey["cpu"].MetricName != "请求CPU执行平均耗时" || byKey["cpu"].ResourceKind != "CPU执行耗时" {
		t.Fatalf("expected request-level CPU resource spec, got %#v", byKey["cpu"])
	}
	if !strings.Contains(byKey["cpu"].Expression, "kindling_profiling_cpu_duration_nanoseconds_sum") ||
		!strings.Contains(byKey["cpu"].Expression, "kindling_span_trace_duration_nanoseconds_count") {
		t.Fatalf("expected CPU resource expression to divide CPU duration by request count, got %s", byKey["cpu"].Expression)
	}

	targetText := requestPolarisTargetText("订单下单接口", "order-service", "POST /orders")
	if targetText != "订单下单接口（服务 order-service，请求路径 POST /orders）" {
		t.Fatalf("expected user-facing request Polaris target text, got %q", targetText)
	}
}

func TestBuildRequestPolarisBreakdownEvidenceIncludesLatencyAndResources(t *testing.T) {
	netChange := 25.0
	tpsChange := 50.0
	trends := []requestPolarisMetricTrend{
		{
			Key:          "latency",
			MetricName:   "请求平均耗时",
			ResourceKind: "请求耗时",
			Unit:         "ms/request",
			PointCount:   3,
			Stats: alertMetricTrendStats{
				FirstValue:   80,
				LastValue:    120,
				AverageValue: 100,
				MaxValue:     140,
				Delta:        40,
			},
			Series: []alertMetricSeries{
				{Values: []prometheus.Points{
					{TimeStamp: 1000, Value: 80},
					{TimeStamp: 2000, Value: 100},
					{TimeStamp: 3000, Value: 120},
				}},
			},
		},
		{
			Key:          "net",
			MetricName:   "请求网络平均耗时",
			ResourceKind: "网络耗时",
			Unit:         "ms/request",
			PointCount:   3,
			Stats: alertMetricTrendStats{
				AverageValue:  45,
				MaxValue:      80,
				Delta:         20,
				ChangePercent: &netChange,
			},
			Abnormal: true,
			Reason:   "网络耗时升高约 25%",
			Series: []alertMetricSeries{
				{Values: []prometheus.Points{
					{TimeStamp: 1000, Value: 35},
					{TimeStamp: 2000, Value: 45},
					{TimeStamp: 3000, Value: 55},
				}},
			},
		},
		{
			Key:          "cpu",
			MetricName:   "请求CPU执行平均耗时",
			ResourceKind: "CPU执行耗时",
			Unit:         "ms/request",
			PointCount:   3,
			Stats: alertMetricTrendStats{
				AverageValue: 30,
				MaxValue:     45,
			},
			Series: []alertMetricSeries{
				{Values: []prometheus.Points{
					{TimeStamp: 1000, Value: 25},
					{TimeStamp: 2000, Value: 30},
					{TimeStamp: 3000, Value: 35},
				}},
			},
		},
		{
			Key:          "epoll",
			MetricName:   "请求调度等待平均耗时",
			ResourceKind: "调度等待耗时",
			Unit:         "ms/request",
			PointCount:   3,
			Stats: alertMetricTrendStats{
				AverageValue: 12,
				MaxValue:     16,
			},
			Series: []alertMetricSeries{
				{Values: []prometheus.Points{
					{TimeStamp: 1000, Value: 10},
					{TimeStamp: 2000, Value: 12},
					{TimeStamp: 3000, Value: 14},
				}},
			},
		},
		{
			Key:          "tps",
			MetricName:   "请求吞吐量",
			ResourceKind: "TPS",
			Unit:         "req/s",
			PointCount:   3,
			Stats: alertMetricTrendStats{
				FirstValue:    10,
				LastValue:     15,
				AverageValue:  12,
				MaxValue:      15,
				Delta:         5,
				ChangePercent: &tpsChange,
			},
		},
	}

		startTime := time.Unix(100, 0)
		endTime := time.Unix(200, 0)
		breakdown := buildRequestPolarisBreakdownEvidence("order-service / POST /orders", "order-service", "POST /orders", startTime, endTime, trends)

		if breakdown == nil || breakdown.Target != "order-service / POST /orders" {
			t.Fatalf("expected target breakdown, got %#v", breakdown)
		}
		if breakdown.StartTime != startTime.UnixMicro() || breakdown.EndTime != endTime.UnixMicro() {
			t.Fatalf("expected breakdown window %d-%d, got %d-%d", startTime.UnixMicro(), endTime.UnixMicro(), breakdown.StartTime, breakdown.EndTime)
		}
	if breakdown.Latency == nil || breakdown.Latency.Name != "请求平均耗时" || breakdown.Latency.AverageValue != 100 || breakdown.Latency.MaxValue != 140 {
		t.Fatalf("expected request latency item, got %#v", breakdown.Latency)
	}
	if len(breakdown.Resources) != 3 {
		t.Fatalf("expected three resource items, got %#v", breakdown.Resources)
	}
	if breakdown.Resources[0].Key != "net" || !breakdown.Resources[0].Abnormal || breakdown.Resources[0].Reason == "" {
		t.Fatalf("expected first resource item to be abnormal net duration, got %#v", breakdown.Resources[0])
	}
	if breakdown.Resources[1].Key != "cpu" || breakdown.Resources[2].Key != "epoll" {
		t.Fatalf("expected resource order net/cpu/epoll, got %#v", breakdown.Resources)
	}
	if len(breakdown.Latency.Points) != 3 || breakdown.Latency.Points[0].Timestamp != 1000 || breakdown.Latency.Points[0].Value != 80 {
		t.Fatalf("expected latency chart points, got %#v", breakdown.Latency.Points)
	}
	if len(breakdown.Resources[0].Points) != 3 || breakdown.Resources[0].Points[2].Value != 55 {
		t.Fatalf("expected resource chart points, got %#v", breakdown.Resources[0].Points)
	}
	if breakdown.TPS == nil || breakdown.TPS.Unit != "req/s" || breakdown.TPS.LastValue != 15 {
		t.Fatalf("expected TPS background item, got %#v", breakdown.TPS)
	}
}

func TestRequestPolarisTrendAbnormalReason(t *testing.T) {
	tpsIncrease := 25.0
	netIncrease := 21.0
	negativeIncrease := 80.0
	cases := []struct {
		name           string
		trend          requestPolarisMetricTrend
		wantAbnormal   bool
		wantReasonPart string
	}{
		{
			name:         "empty trend",
			trend:        requestPolarisMetricTrend{Key: "net"},
			wantAbnormal: false,
		},
		{
			name: "tps increase is abnormal but not resource root cause",
			trend: requestPolarisMetricTrend{
				Key:        "tps",
				PointCount: 3,
				Stats: alertMetricTrendStats{
					Delta:         5,
					ChangePercent: &tpsIncrease,
				},
			},
			wantAbnormal:   true,
			wantReasonPart: "TPS 升高约",
		},
		{
			name: "latency is reported but not resource abnormal",
			trend: requestPolarisMetricTrend{
				Key:        "latency",
				PointCount: 3,
				Stats: alertMetricTrendStats{
					MaxValue: 120,
				},
			},
			wantAbnormal: false,
		},
		{
			name: "resource max value crosses threshold",
			trend: requestPolarisMetricTrend{
				Key:          "net",
				ResourceKind: "网络耗时",
				PointCount:   3,
				Stats: alertMetricTrendStats{
					MaxValue: requestPolarisResourceThresholdMS,
				},
			},
			wantAbnormal:   true,
			wantReasonPart: "网络耗时最高达到",
		},
		{
			name: "resource increase crosses threshold",
			trend: requestPolarisMetricTrend{
				Key:          "epoll",
				ResourceKind: "调度等待耗时",
				PointCount:   3,
				Stats: alertMetricTrendStats{
					Delta:         4,
					ChangePercent: &netIncrease,
				},
			},
			wantAbnormal:   true,
			wantReasonPart: "调度等待耗时升高约",
		},
		{
			name: "increase percent without positive delta is ignored",
			trend: requestPolarisMetricTrend{
				Key:          "net",
				ResourceKind: "网络耗时",
				PointCount:   3,
				Stats: alertMetricTrendStats{
					Delta:         -1,
					MaxValue:      10,
					ChangePercent: &negativeIncrease,
				},
			},
			wantAbnormal: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotAbnormal, gotReason := requestPolarisTrendAbnormalReason(tc.trend)
			if gotAbnormal != tc.wantAbnormal {
				t.Fatalf("abnormal=%v, want %v, reason=%q", gotAbnormal, tc.wantAbnormal, gotReason)
			}
			if tc.wantReasonPart == "" {
				if gotReason != "" {
					t.Fatalf("expected empty reason, got %q", gotReason)
				}
				return
			}
			if !strings.Contains(gotReason, tc.wantReasonPart) {
				t.Fatalf("expected reason to contain %q, got %q", tc.wantReasonPart, gotReason)
			}
		})
	}
}

func TestBuildRequestPolarisEvidenceResultSummarizesAbnormalTrends(t *testing.T) {
	netChange := 25.0
	tpsChange := 50.0
	trends := []requestPolarisMetricTrend{
		{
			Key:          "net",
			ResourceKind: "网络耗时",
			PointCount:   3,
			Stats: alertMetricTrendStats{
				MaxValue:      80,
				ChangePercent: &netChange,
			},
		},
		{
			Key:          "epoll",
			ResourceKind: "调度等待耗时",
			PointCount:   3,
			Stats: alertMetricTrendStats{
				MaxValue: 60,
			},
		},
		{
			Key:        "tps",
			PointCount: 3,
			Stats: alertMetricTrendStats{
				FirstValue:    10,
				LastValue:     15,
				Delta:         5,
				ChangePercent: &tpsChange,
			},
		},
	}

	result := buildRequestPolarisEvidenceResult(
		"order-service / POST /orders",
		"order-service",
		"POST /orders",
		trends,
		trends[:2],
		nil,
		0,
		nil,
	)

	for _, expected := range []string{
		"请求北极星显示",
		"order-service / POST /orders",
		"网络耗时异常",
		"最高 80",
		"窗口变化约 +25",
		"同窗口 TPS 从 10",
		"整体升高",
		"变化约 +50",
		"同时可见调度等待耗时最高 60",
	} {
		if !strings.Contains(result, expected) {
			t.Fatalf("expected %q in result: %s", expected, result)
		}
	}
}

func TestBuildRequestPolarisEvidenceResultSummarizesFallbacks(t *testing.T) {
	tpsChange := -20.0
	trends := []requestPolarisMetricTrend{
		{
			Key:        "tps",
			PointCount: 3,
			Stats: alertMetricTrendStats{
				FirstValue:    20,
				LastValue:     16,
				Delta:         -4,
				ChangePercent: &tpsChange,
			},
		},
	}

	errResult := buildRequestPolarisEvidenceResult(
		"order-service / POST /orders",
		"order-service",
		"POST /orders",
		trends,
		nil,
		nil,
		0,
		errors.New("clickhouse unavailable"),
	)
	for _, expected := range []string{"资源分项未见超过 50ms/请求或明显升高", "慢调用样本检索失败", "clickhouse unavailable", "整体回落"} {
		if !strings.Contains(errResult, expected) {
			t.Fatalf("expected %q in slow-span error result: %s", expected, errResult)
		}
	}

	slowSpanResult := buildRequestPolarisEvidenceResult(
		"order-service / POST /orders",
		"order-service",
		"POST /orders",
		trends,
		nil,
		[]clickhouse.QueryTraceResult{
			{Duration: 800000, IsSlow: true, IsError: true, TraceId: "trace-1", SpanId: "span-1"},
			{Duration: 600000, TraceId: "trace-2", SpanId: "span-2"},
		},
		7,
		nil,
	)
	for _, expected := range []string{"检索到 7 条超过 500ms 的慢调用样本", "本页2条样本中 error=1、slow=1、最大耗时=800.00ms、代表trace=trace-1、span=span-1", "不代表整体资源耗时已异常"} {
		if !strings.Contains(slowSpanResult, expected) {
			t.Fatalf("expected %q in slow-span result: %s", expected, slowSpanResult)
		}
	}

	noAbnormalResult := buildRequestPolarisEvidenceResult(
		"order-service / POST /orders",
		"order-service",
		"",
		trends,
		nil,
		nil,
		0,
		nil,
	)
	for _, expected := range []string{"未检索到超过 500ms 的慢调用样本", "当前不把资源分项耗时作为主要根因"} {
		if !strings.Contains(noAbnormalResult, expected) {
			t.Fatalf("expected %q in no-abnormal result: %s", expected, noAbnormalResult)
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

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func assertNoUserFacingMetricDebugContent(t *testing.T, text string) {
	t.Helper()
	for _, forbidden := range []string{"PQL", "PromQL", "rate(", "sum by(", "query 表达式", "查询表达式"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("user-facing metric text should not expose %q: %s", forbidden, text)
		}
	}
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
