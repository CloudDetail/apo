// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"time"
)

// AlertEvent With Workflow record
type AEventWithWRecord struct {
	AlertEvent

	WorkflowRunID string `json:"workflowRunId" ch:"workflow_run_id"`

	WorkflowID   string `json:"workflowId" ch:"workflow_id"`
	WorkflowName string `json:"workflowName" ch:"workflow_name"`

	Validity    string    `json:"validity" ch:"validity"`
	Output      string    `json:"output" ch:"output"`
	RoundedTime time.Time `json:"-" ch:"rounded_time"`
	Importance  uint8     `json:"-" ch:"importance"`

	WorkflowParams WorkflowParams `json:"workflowParams"`

	// Deprecated: use [Validity] instead, will remove after 1.7.x
	IsValid string `json:"isValid" ch:"is_valid"`
}

type WorkflowParams struct {
	StartTime int64 `json:"startTime" form:"startTime"`
	EndTime   int64 `json:"endTime" form:"endTime"`

	NodeName string `json:"nodeName" form:"nodeName"`
	NodeIp   string `json:"nodeIp" form:"nodeIp"`

	Params string `json:"params"`
}

type AlertAnalyzeWorkflowParams struct {
	Node      string `json:"node,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Pod       string `json:"pod,omitempty"`
	Service   string `json:"service,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	Pid       string `json:"pid,omitempty"`
	AlertName string `json:"alertName,omitempty"`
}
