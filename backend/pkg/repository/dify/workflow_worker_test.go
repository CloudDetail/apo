// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func TestWorkflowParamsIncludeReportOutputContract(t *testing.T) {
	w := &worker{AlertCheckConfig: &AlertCheckConfig{}}
	event := &alert.AlertEvent{
		EventID:    "event-001",
		UpdateTime: time.Unix(1779258240, 0),
		Alert: alert.Alert{
			Name: "CheckoutErrorRateHigh",
			EnrichTags: map[string]string{
				"serviceName": "checkout",
				"endpoint":    "/api/checkout/submit",
			},
			Tags: alert.RawTags{
				"source": "prometheus",
			},
		},
	}

	params := w.getWorkflowParams(event)

	var payload alert.AlertAnalyzeWorkflowParams
	if err := json.Unmarshal([]byte(params.Params), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ReportOutputContract == nil {
		t.Fatalf("reportOutputContract missing from workflow params: %s", params.Params)
	}
	if !strings.Contains(payload.ReportOutputContract.OverviewDetail, "异常症状") ||
		!strings.Contains(payload.ReportOutputContract.FirstScreenConstraint, "三个诊断卡片") ||
		!strings.Contains(payload.ReportOutputContract.Suggest, "短标题") ||
		!strings.Contains(payload.ReportOutputContract.Evidence, "major") ||
		!strings.Contains(payload.ReportOutputContract.Evidence, "RED") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "增量输出") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "不要修改已有") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "rootCauseEvidenceRefs") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "首屏支撑证据唯一来源") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "metrics/red/log/traces") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "type 别名归一") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "disk 按 storage") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "file 按 runtime/storage") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "memory 按 runtime/infra") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "system 按 infra/runtime/cpu/memory") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "config 按 config/dependency/log") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "self 按 code/log/runtime") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "rootCauseMetricRefs") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "兼容") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "evidenceHighlights 只服务关键证据摘要") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "timeline 使用 timestamp/title/summary") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "analysisSteps 使用 title/summary") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "下游服务") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "MySQL") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "锁等待") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "连接池等待") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "配置拉取失败") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "DB 连接失败") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "OOMKill") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "文件描述符") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "Read timed out") ||
		!strings.Contains(payload.ReportOutputContract.ReportView, "正常或已排除证据") {
		t.Fatalf("reportOutputContract is not the first-screen contract: %#v", payload.ReportOutputContract)
	}
}

func TestRunAlertAnalyzeWithRetryRetriesTransientFailures(t *testing.T) {
	attempts := 0
	waits := 0

	resp, err := runAlertAnalyzeWithRetry(
		func() (*AlertAnalyzeResponse, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("temporary dify api failure")
			}
			return &AlertAnalyzeResponse{}, nil
		},
		3,
		func(time.Duration) { waits++ },
	)
	if err != nil {
		t.Fatalf("runAlertAnalyzeWithRetry() error = %v", err)
	}
	if resp == nil {
		t.Fatal("runAlertAnalyzeWithRetry() returned nil response")
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	if waits != 2 {
		t.Fatalf("waits = %d, want 2", waits)
	}
}

func TestDefaultAlertClassifyUsesAnalyzeWorkflowConfig(t *testing.T) {
	w := &worker{AlertCheckConfig: &AlertCheckConfig{
		FlowId:        "alert-check-flow",
		AnalyzeFlowId: "alert-analyze-flow",
		AnalyzeAPIKey: "app-analyze",
		AnalyzeAuth:   "Bearer app-analyze",
	}}

	got := w.defaultAlertClassify()
	if got.WorkflowId != "alert-analyze-flow" {
		t.Fatalf("WorkflowId = %q, want alert-analyze-flow", got.WorkflowId)
	}
	if got.WorkflowApiKey != "app-analyze" {
		t.Fatalf("WorkflowApiKey = %q, want app-analyze", got.WorkflowApiKey)
	}
}
