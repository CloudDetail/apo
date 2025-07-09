// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

type WorkflowRecord struct {
	WorkflowRunID string `json:"workflowRunId" ch:"workflow_run_id"`

	WorkflowID   string `json:"workflowId" ch:"workflow_id"`
	WorkflowName string `json:"workflowName" ch:"workflow_name"`

	Ref    string `json:"ref" ch:"ref"`
	Input  string `json:"input" ch:"input"`
	Output string `json:"output" ch:"output"`

	CreatedAt   int64 `json:"createdAt" ch:"created_at"`
	RoundedTime int64 `json:"-" ch:"rounded_time"`

	InputRef any `json:"-" ch:"-"`

	AlertDirection string `json:"alertDirection" ch:"alert_direction"`
	AnalyzeRunID   string `json:"analyzeRunId" ch:"analyze_run_id"`
	AnalyzeErr     string `json:"analyzeErr" ch:"analyze_err"`
}

type AlertNotifyRecord struct {
	AlertID   string `json:"alertId" ch:"alert_id"`
	CreatedAt int64  `json:"createdAt" ch:"created_at"`
	EventID   string `json:"eventId" ch:"event_id"`
	Success   string `json:"success" ch:"success"`
	Failed    string `json:"failed" ch:"failed"`
}

type AlertEventClassify struct {
	WorkflowId     string `json:"workflowId"`
	WorkflowApiKey string `json:"workflowApiKey"`
}
