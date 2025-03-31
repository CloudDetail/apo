// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DifyClient struct {
	URL string

	client http.Client
}

func NewDifyClient(client *http.Client, url string) *DifyClient {
	if !strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	return &DifyClient{
		URL:    url,
		client: *client,
	}
}

func (c *DifyClient) alertCheck(req *DifyRequest, authorization string, user string) (*AlertCheckResponse, error) {
	req.ResponseMode = "blocking"
	resp, err := c.WorkflowsRun(req, authorization, user)
	if err != nil {
		return nil, err
	}
	if resp, ok := resp.(*CompletionResponse); ok {
		return &AlertCheckResponse{resp}, err
	}
	return nil, fmt.Errorf("alertCheck must be run in blocking mode")
}

func (c *DifyClient) WorkflowsRun(req *DifyRequest, authorization string, user string) (DifyResponse, error) {
	req.User = user
	jsonBytes, _ := json.Marshal(req)
	fullReq, _ := http.NewRequest("POST", c.URL+"/v1/workflows/run", bytes.NewReader(jsonBytes))

	fullReq.Header.Set("Content-Type", "application/json")
	fullReq.Header.Set("Authorization", authorization)

	resp, err := c.client.Do(fullReq)
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

type DifyRequest struct {
	Inputs       json.RawMessage `json:"inputs"`
	ResponseMode string          `json:"response_mode"`
	User         string          `json:"user"`
}

type DifyResponse interface {
	_DifyResponse()
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

func (r *CompletionResponse) _DifyResponse() {}

type ChunkCompletionResponse struct{}

func (r *ChunkCompletionResponse) _DifyResponse() {}

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

	return text
}
