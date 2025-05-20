// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"go.uber.org/zap"
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

	WorkflowsRun(req *WorkflowRequest, authorization string) (*CompletionResponse, error)
}

type difyRepo struct {
	cli *DifyClient
	url string

	asyncAlertCheck

	AlertCheckCFG      AlertCheckConfig
	AlertAnalyzeFlowId string
}

func New() (DifyRepo, error) {
	// client := &http.Client{}
	difyConf := config.Get().Dify
	return &difyRepo{
		cli: &DifyClient{
			Client:  DefaultDifyFastHttpClient,
			BaseURL: difyConf.URL,
		},
		AlertCheckCFG: DefaultAlertCheckConfig(),
		url:           difyConf.URL,
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

func (r *difyRepo) WorkflowsRun(req *WorkflowRequest, authorization string) (*CompletionResponse, error) {
	resp, err := r.cli.WorkflowsRun(req, authorization)
	if completResp, ok := resp.(*CompletionResponse); ok {
		return completResp, err
	}
	return nil, fmt.Errorf("only support block request now")
}
