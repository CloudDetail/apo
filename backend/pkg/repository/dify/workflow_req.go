// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"encoding/json"
	"fmt"
	"strings"
)

type WorkflowRequest struct {
	Inputs       json.RawMessage `json:"inputs"`
	ResponseMode string          `json:"response_mode"`
	User         string          `json:"user"`
}

type WorkflowResponse interface {
	_WorkflowResponse()
}

type CompletionResponse struct {
	WorkflowRunID string                 `json:"workflow_run_id"`
	TaskID        string                 `json:"task_id"`
	Data          CompletionResponseData `json:"data"`
}

type CompletionResponseData struct {
	ID         string          `json:"id"`
	WorkflowID string          `json:"workload_id"`
	Status     string          `json:"status"`
	Outputs    json.RawMessage `json:"outputs"`

	// Optional Response
	// Error      string          `json:"error,omitempty"`
	// ...

	CreatedAt int64 `json:"created_at"`
}

func (r *CompletionResponse) _WorkflowResponse() {}

type ChunkCompletionResponse struct{}

func (r *ChunkCompletionResponse) _WorkflowResponse() {}

type AlertAnalyzeResponse struct {
	resp *CompletionResponse
}

type AlertCheckResponse struct {
	resp *CompletionResponse
}

func (r *AlertCheckResponse) WorkflowRunID() string {
	return r.resp.WorkflowRunID
}

// UnixMicro Timestamp
func (r *AlertCheckResponse) CreatedAt() int64 {
	return r.resp.Data.CreatedAt * 1e6
}

func (r *AlertCheckResponse) getOutput(defaultV string) string {
	if r.resp.Data.Status != "succeeded" {
		return fmt.Sprintf("failed: status: %s, output: %s", r.resp.Data.Status, string(r.resp.Data.Outputs))
	}

	var res map[string]string
	err := json.Unmarshal(r.resp.Data.Outputs, &res)
	if err != nil {
		return defaultV
	}

	text, find := res["text"]
	if !find {
		return defaultV
	}

	return normalizeAlertCheckOutput(text)
}

func normalizeAlertCheckOutput(text string) string {
	cleaned := stripThinkBlocks(text)
	switch strings.ToLower(cleaned) {
	case "true":
		return "true"
	case "false":
		return "false"
	default:
		return text
	}
}

func stripThinkBlocks(text string) string {
	for {
		lower := strings.ToLower(text)
		start := strings.Index(lower, "<think>")
		if start < 0 {
			return strings.TrimSpace(text)
		}
		end := strings.Index(lower[start:], "</think>")
		if end < 0 {
			return strings.TrimSpace(text[:start])
		}
		text = text[:start] + text[start+end+len("</think>"):]
	}
}

func (r *AlertAnalyzeResponse) WorkflowRunID() string {
	return r.resp.WorkflowRunID
}

func (r *AlertAnalyzeResponse) getOutput(defaultV string) string {
	if r.resp.Data.Status != "succeeded" {
		// return fmt.Sprintf("failed: status: %s, output: %s", r.resp.Data.Status, string(r.resp.Data.Outputs))
		return "workflow execution failed"
	}

	var res map[string]string
	err := json.Unmarshal(r.resp.Data.Outputs, &res)
	if err != nil {
		return defaultV
	}

	text, find := res["alertDirection"]
	if !find {
		return defaultV
	}

	return stripThinkBlocks(text)
}
