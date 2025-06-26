// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"time"
)

// AlertEvent With Workflow record
type AEventWithWRecord struct {
	AlertEvent

	LastStatus string `json:"lastStatus,omitempty" ch:"last_status"`

	WorkflowDetail
	NotifyDetail
}

// FIXME add importance field
type WorkflowDetail struct {
	WorkflowRunID string `json:"workflowRunId" ch:"workflow_run_id"`

	WorkflowID   string `json:"workflowId" ch:"workflow_id"`
	WorkflowName string `json:"workflowName" ch:"workflow_name"`

	Validity    string    `json:"validity" ch:"validity"`
	Output      string    `json:"output" ch:"output"`
	RoundedTime time.Time `json:"-" ch:"rounded_time"`
	Importance  uint8     `json:"-" ch:"importance"`
	LastCheckAt time.Time `json:"lastCheckAt" ch:"last_check_at"`

	Duration string `json:"duration" ch:"-"`

	WorkflowParams WorkflowParams `json:"workflowParams"`

	// Deprecated: use [Validity] instead, will remove after 1.7.x
	IsValid string `json:"isValid" ch:"is_valid"`

	AlertDirection string `json:"alertDirection" ch:"alert_direction"`
	AnalyzeRunID   string `json:"analyzeRunId" ch:"analyze_run_id"`
	AnalyzeErr     string `json:"analyzeErr" ch:"analyze_err"`
}

type NotifyDetail struct {
	SendSuccess string    `json:"notifySuccess" ch:"notify_success"`
	SendFailed  string    `json:"notifyFailed" ch:"notify_failed"`
	NotifyAt    time.Time `json:"notifyAt" ch:"notify_at"`
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

	Detail      string `json:"detail,omitempty"`
	ContainerID string `json:"containerId,omitempty"`

	Tags    map[string]string `json:"tags,omitempty"`
	RawTags map[string]any    `json:"raw_tags,omitempty"`
}
