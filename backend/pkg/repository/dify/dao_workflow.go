// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (d *difyRepo) WorkflowsRun(req *DifyWorkflowRequest, authorization string) (*CompletionResponse, error) {
	jsonBytes, _ := json.Marshal(req)
	fullReq, _ := http.NewRequest("POST", d.url+"/v1/workflows/run", bytes.NewReader(jsonBytes))

	fullReq.Header.Set("Content-Type", "application/json")
	fullReq.Header.Set("Authorization", authorization)

	resp, err := d.cli.Do(fullReq)
	if err != nil {
		return nil, fmt.Errorf("failed to run workflow, err: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var bodyStr = ""
		if body, err := io.ReadAll(resp.Body); err == nil {
			bodyStr = string(body)
		}
		return nil, fmt.Errorf("failed to run workflow, [%d] %s", resp.StatusCode, bodyStr)
	}

	switch req.ResponseMode {
	case "blocking":
		var completionResponse CompletionResponse
		err = json.NewDecoder(resp.Body).Decode(&completionResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse completion response, err: %w", err)
		}
		return &completionResponse, nil
	case "streaming":
		panic("not implemented yet")
	}

	return nil, nil
}
