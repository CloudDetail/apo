// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"encoding/json"
	"net/http"

	"github.com/CloudDetail/apo/backend/config"
)

const (
	DIFY_WORKFLOWS_RUN   = "/v1/workflows/run"
	DIFY_ADD_USER        = "/console/api/workspaces/apo/members/add"
	DIFY_REMOVE_USER     = "/console/api/workspaces/apo/members/"
	DIFY_PASSWORD_UPDATE = "/console/api/apo/account/password"
	DIFY_RESET_PASSWORD  = "/console/api/apo/account/reset-password"
)

type DifyUser struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
	Role        string `json:"role"`
	Username    string `json:"username"`
}

type DifyResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

type DifyWorkflowRequest struct {
	Inputs       json.RawMessage `json:"inputs"`
	ResponseMode string          `json:"response_mode"`
	User         string          `json:"user"`
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

type DifyRepo interface {
	AddUser(username string, password string, role string) (*DifyResponse, error)
	UpdatePassword(username string, oldPassword string, newPassword string) (*DifyResponse, error)
	RemoveUser(username string) (*DifyResponse, error)
	ResetPassword(username string, newPassword string) (*DifyResponse, error)
	WorkflowsRun(req *DifyWorkflowRequest, authorization string) (*CompletionResponse, error)
}

type difyRepo struct {
	cli *http.Client
	url string
}

func New() (DifyRepo, error) {
	client := &http.Client{}
	difyConf := config.Get().Dify
	return &difyRepo{
		cli: client,
		url: difyConf.URL,
	}, nil
}
