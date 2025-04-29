// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"go.uber.org/zap"
	"encoding/json"
	"net/http"

	"github.com/CloudDetail/apo/backend/config"
)

const (
	DIFY_WORKFLOWS_RUN = "/v1/workflows/run"
	DIFY_ADD_USER      = "/console/api/workspaces/apo/members/add"
	DIFY_REMOVE_USER   = "/console/api/workspaces/apo/members/"
	DIFY_PASSWD_UPDATE = "/console/api/apo/account/password"
	DIFY_RESET_PASSWD  = "/console/api/apo/account/reset-password"
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

	// ########################## Async AlertCheck Workflow #########################

	PrepareAsyncAlertCheckWorkflow(cfg *AlertCheckConfig, logger *zap.Logger) (records <-chan *model.WorkflowRecord, err error)
	SubmitAlertEvents(events []alert.AlertEvent)

	GetCacheMinutes() int
	GetAlertCheckFlowID() string
	GetAlertAnalyzeFlowID() string
}

type difyRepo struct {
	cli *DifyClient

	asyncAlertCheck

	AlertCheckCFG      AlertCheckConfig
	AlertAnalyzeFlowId string
}

func New() (DifyRepo, error) {
	// client := &http.Client{}
	return &difyRepo{
		cli:           &DifyClient{},
		AlertCheckCFG: DefaultAlertCheckConfig(),
	}, nil
}

func (r *difyRepo) GetCacheMinutes() int {
	return r.AlertCheckCFG.CacheMinutes
}

func (r *difyRepo) GetAlertCheckFlowID() string {
	return r.AlertCheckCFG.FlowId
}

func (r *difyRepo) GetAlertAnalyzeFlowID() string {
	return r.AlertAnalyzeFlowId
}
