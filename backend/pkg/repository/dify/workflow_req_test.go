// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"encoding/json"
	"testing"
)

func TestAlertCheckResponseGetOutputStripsThinkBlock(t *testing.T) {
	outputs, err := json.Marshal(map[string]string{
		"text": "<think>reasoning that mentions true and false</think>false",
	})
	if err != nil {
		t.Fatal(err)
	}

	resp := &AlertCheckResponse{resp: &CompletionResponse{
		Data: CompletionResponseData{
			Status:  "succeeded",
			Outputs: outputs,
		},
	}}

	got := resp.getOutput("default")
	if got != "false" {
		t.Fatalf("getOutput() = %q, want false", got)
	}
}

func TestAlertAnalyzeResponseGetOutputStripsThinkBlock(t *testing.T) {
	outputs, err := json.Marshal(map[string]string{
		"alertDirection": "<think>reasoning about the network</think>网络异常",
	})
	if err != nil {
		t.Fatal(err)
	}

	resp := &AlertAnalyzeResponse{resp: &CompletionResponse{
		Data: CompletionResponseData{
			Status:  "succeeded",
			Outputs: outputs,
		},
	}}

	got := resp.getOutput("default")
	if got != "网络异常" {
		t.Fatalf("getOutput() = %q, want 网络异常", got)
	}
}
