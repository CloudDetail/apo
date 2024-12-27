// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
)

var _ Service = (*service)(nil)

type Service interface {
	// InputAlertManager 接收 AlertManager 的告警事件
	InputAlertManager(req *request.InputAlertManagerRequest) error
	ForwardToDingTalk(req *request.ForwardToDingTalkRequest, uuid string) error

	// GetAlertRuleFile 获取基础告警规则
	GetAlertRuleFile(req *request.GetAlertRuleConfigRequest) (*response.GetAlertRuleFileResponse, error)
	// UpdateAlertRuleFile 更新告警基础规则
	UpdateAlertRuleFile(req *request.UpdateAlertRuleConfigRequest) error

	// AlertRule Options
	GetGroupList() response.GetGroupListResponse
	GetMetricPQL() (*response.GetMetricPQLResponse, error)

	// AlertRule CRUD
	GetAlertRules(req *request.GetAlertRuleRequest) response.GetAlertRulesResponse
	UpdateAlertRule(req *request.UpdateAlertRuleRequest) error
	DeleteAlertRule(req *request.DeleteAlertRuleRequest) error
	AddAlertRule(req *request.AddAlertRuleRequest) error
	CheckAlertRule(req *request.CheckAlertRuleRequest) (response.CheckAlertRuleResponse, error)

	// AlertManager Receiver CRUD
	GetAMConfigReceivers(req *request.GetAlertManagerConfigReceverRequest) response.GetAlertManagerConfigReceiverResponse
	AddAMConfigReceiver(req *request.AddAlertManagerConfigReceiver) error
	UpdateAMConfigReceiver(req *request.UpdateAlertManagerConfigReceiver) error
	DeleteAMConfigReceiver(req *request.DeleteAlertManagerConfigReceiverRequest) error
}

type service struct {
	chRepo clickhouse.Repo
	k8sApi kubernetes.Repo
	dbRepo database.Repo
}

func New(chRepo clickhouse.Repo, k8sApi kubernetes.Repo, dbRepo database.Repo) Service {
	return &service{
		chRepo: chRepo,
		k8sApi: k8sApi,
		dbRepo: dbRepo,
	}
}
