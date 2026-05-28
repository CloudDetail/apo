package response

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func TestToAlertReportViewResponseRestoresReportViewFromHiddenTag(t *testing.T) {
	reportView := map[string]any{
		"headline":          "应用错误日志异常增长，更可能导致 CPU 异常",
		"conclusionSummary": "CPU 使用率与错误日志数在告警窗口同步升高。",
		"rootCauseMetricRefs": []any{
			map[string]any{"evidenceIndex": float64(0), "title": "CPU 使用率"},
		},
	}
	reportViewBytes, err := json.Marshal(reportView)
	if err != nil {
		t.Fatal(err)
	}

	report := model.AlertEventReport{
		ReportType:        "infra",
		OverviewTimestamp: time.UnixMicro(1779258240000000),
		OverviewDetail:    "错误日志数升高",
		OverviewReason:    "CPU 与错误日志同步升高",
		OverviewTags: map[string]string{
			"service":                          "checkout",
			model.ReportViewTagKey:             string(reportViewBytes),
			model.ReportOverviewHeadlineTagKey: "legacy headline",
		},
		RootCause:      "CPU 方向异常",
		RootCauseOther: []string{`{"level":"major","type":"cpu","reason":"CPU 与错误日志同步升高"}`},
		Data:           []string{`{"name":"CPU 使用率","type":"metrics","data":[{"legend":"cpu usage","chart":{"chartData":{"1779256980000000":45,"1779258240000000":96}}}]}`},
	}

	resp := ToAlertReportViewResponse(report, &request.GetAlertEventReportRequest{
		StartTime: 1779256980000000,
		EndTime:   1779258240000000,
		Tags:      map[string]string{"alertEventId": "event-001"},
	})

	data := resp.Snapshot.Data.(ReportData)
	if data.ReportView == nil {
		t.Fatalf("snapshot reportView is nil")
	}
	if got := data.ReportView["headline"]; got != reportView["headline"] {
		t.Fatalf("reportView headline = %#v, want %#v", got, reportView["headline"])
	}
	if _, ok := data.OverView.Tags[model.ReportViewTagKey]; ok {
		t.Fatalf("public overview tags leaked %s", model.ReportViewTagKey)
	}
	if _, ok := data.OverView.Tags[model.ReportOverviewHeadlineTagKey]; ok {
		t.Fatalf("public overview tags leaked %s", model.ReportOverviewHeadlineTagKey)
	}
}

func TestToAlertReportViewResponseDoesNotDeriveDisplayViewFromSnapshotText(t *testing.T) {
	report := model.AlertEventReport{
		ReportType:        "error",
		OverviewTimestamp: time.UnixMicro(1779258240000000),
		OverviewDetail:    "checkout 下单接口错误率升高",
		OverviewReason:    "14:03 checkout 错误率开始升高，告警触发。",
		OverviewTags: map[string]string{
			"service":  "checkout",
			"endpoint": "/api/checkout/submit",
		},
		Topology: `{
			"current":{"service":"checkout","description":"下单入口服务"},
			"parents":[{"service":"web-bff","description":"用户入口"}],
			"children":[{"service":"payment-gateway","description":"第三方支付网关"}]
		}`,
		RootCause:      "payment-gateway 第三方 5xx / 超时导致 checkout 下单失败",
		RootCauseOther: []string{`{"level":"major","type":"downstream","reason":"payment-gateway 5xx 同步升高"}`},
		Suggest:        []string{`{"action":"切换备用支付通道","level":"high"}`},
		Data:           []string{`{"name":"checkout 错误率","type":"metrics","description":"错误率升高","unit":"percent","data":[{"legend":"checkout error rate","chart":{"chartData":{"1779256980000000":0.6,"1779258240000000":4.8}}}]}`},
	}

	resp := ToAlertReportViewResponse(report, &request.GetAlertEventReportRequest{
		StartTime: 1779256980000000,
		EndTime:   1779258240000000,
		Tags:      map[string]string{"alertEventId": "event-001"},
	})

	if resp.View.Headline != "" {
		t.Fatalf("view headline = %q, want no derived headline", resp.View.Headline)
	}
	if resp.View.AnalysisWindow != nil {
		t.Fatalf("view analysisWindow = %#v, want nil", resp.View.AnalysisWindow)
	}
	if resp.View.BusinessImpact != nil {
		t.Fatalf("view businessImpact = %#v, want nil", resp.View.BusinessImpact)
	}
	if len(resp.View.MetricDeltas) != 0 {
		t.Fatalf("view metricDeltas = %#v, want empty", resp.View.MetricDeltas)
	}
	if resp.View.FaultLocation != nil {
		t.Fatalf("view faultLocation = %#v, want nil", resp.View.FaultLocation)
	}
	if len(resp.View.Timeline) != 0 {
		t.Fatalf("view timeline = %#v, want empty", resp.View.Timeline)
	}
	if resp.View.Topology != nil {
		t.Fatalf("view topology = %#v, want nil", resp.View.Topology)
	}
	if resp.View.DataSources.Snapshot != "alert_report" {
		t.Fatalf("snapshot source = %q, want alert_report", resp.View.DataSources.Snapshot)
	}
}
