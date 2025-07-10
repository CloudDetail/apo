// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
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

	asyncAlertCheck

	AlertCheckCFG *AlertCheckConfig
}

func New(prom prometheus.Repo, logger *zap.Logger) (DifyRepo, <-chan *WorkflowRecordWithCtx, error) {
	// client := &http.Client{}
	cfg := config.Get().Dify

	if len(cfg.Sampling) == 0 {
		cfg.Sampling = "first"
	}
	if cfg.CacheMinutes <= 0 {
		cfg.CacheMinutes = 20
	} else {
		cfg.CacheMinutes = maxFactorOf60LessThanN(cfg.CacheMinutes)
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 1
	}

	repo, err := newRepo(prom, cfg)
	if err != nil {
		return nil, nil, err
	}
	record, err := repo.PrepareAsyncAlertCheckWorkflow(prom, logger)
	return repo, record, err
}

func newRepo(prom prometheus.Repo, cfg config.DifyConfig) (*difyRepo, error) {
	if cfg.TimeoutSecond <= 0 {
		cfg.TimeoutSecond = 180
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
		Timeout: time.Duration(cfg.TimeoutSecond) * time.Second,
	}

	alertCfg := &AlertCheckConfig{
		DifyConfig:     cfg,
		AlertCheckAuth: fmt.Sprintf("Bearer %s", cfg.APIKeys.AlertCheck),
		AnalyzeAuth:    fmt.Sprintf("Bearer %s", cfg.APIKeys.AlertAnalyze),
		User:           "apo-backend",
		Prom:           prom,
	}

	return &difyRepo{
		cli: &DifyClient{
			Client:  client,
			BaseURL: cfg.URL,
		},
		AlertCheckCFG: alertCfg,
	}, nil
}

func (r *difyRepo) GetCacheMinutes() int {
	return r.AlertCheckCFG.CacheMinutes
}

func (r *difyRepo) GetAlertCheckFlowID() string {
	return r.AlertCheckCFG.FlowIDs.AlertCheck
}

func (r *difyRepo) GetAlertAnalyzeFlowID() string {
	return r.AlertCheckCFG.FlowIDs.AlertEventAnalyze
}

func (r *difyRepo) WorkflowsRun(req *WorkflowRequest, authorization string) (*CompletionResponse, error) {
	resp, err := r.cli.WorkflowsRun(req, authorization)
	if err != nil {
		return nil, err
	}
	if completResp, ok := resp.(*CompletionResponse); ok {
		return completResp, err
	}
	return nil, fmt.Errorf("only support block request now")
}
