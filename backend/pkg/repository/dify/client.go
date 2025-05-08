// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type DifyClient struct {
	*http.Client

	BaseURL string
}

func (c *DifyClient) alertCheck(req *WorkflowRequest, authorization string, user string) (*AlertCheckResponse, error) {
	req.ResponseMode = "blocking"
	req.User = user
	resp, err := c.WorkflowsRun(req, authorization)
	if err != nil {
		return nil, err
	}
	if resp, ok := resp.(*CompletionResponse); ok {
		return &AlertCheckResponse{resp}, err
	}
	return nil, fmt.Errorf("alertCheck must be run in blocking mode")
}

func (c *DifyClient) WorkflowsRun(req *WorkflowRequest, authorization string) (WorkflowResponse, error) {
	jsonBytes, _ := json.Marshal(req)
	fullReq, _ := http.NewRequest("POST", c.BaseURL+"/v1/workflows/run", bytes.NewReader(jsonBytes))

	fullReq.Header.Set("Content-Type", "application/json")
	fullReq.Header.Set("Authorization", authorization)

	resp, err := c.Client.Do(fullReq)
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

var DefaultDifyFastHttpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		DialContext: (&net.Dialer{
			Timeout:   1 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
	Timeout: 3 * time.Minute,
}
