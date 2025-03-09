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

	IsValid     string    `json:"isValid" ch:"is_valid"`
	RoundedTime time.Time `json:"-" ch:"rounded_time"`

	WorkflowParams WorkflowParams `json:"workflowParams"`
}

type WorkflowParams struct {
	StartTime int64 `json:"startTime" form:"startTime"`
	EndTime   int64 `json:"endTime" form:"endTime"`

	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`

	Params string `json:"params"`
}

type AlertAnalyzeWorkflowParams struct {
	Node      string `json:"node"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
}
