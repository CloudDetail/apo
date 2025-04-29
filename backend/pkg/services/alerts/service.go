// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/amreceiver"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	// ========================告警检索========================
	AlertEventList(req *request.AlertEventSearchRequest) (*response.AlertEventSearchResponse, error)
	AlertDetail(req *request.GetAlertDetailRequest) (*response.GetAlertDetailResponse, error)

	AlertEventClassify(req *request.AlertEventClassifyRequest) (*response.AlertEventClassifyResponse, error)

	// ========================告警配置========================

	// InputAlertManager receive AlertManager alarm events
	// Deprecated: use alertinput.ProcessAlertEvents instead
	InputAlertManager(req *request.InputAlertManagerRequest) error
	ForwardToDingTalk(req *request.ForwardToDingTalkRequest, uuid string) error

	// GetAlertRuleFile get basic alarm rules
	GetAlertRuleFile(req *request.GetAlertRuleConfigRequest) (*response.GetAlertRuleFileResponse, error)
	// UpdateAlertRuleFile update basic alarm rules
	UpdateAlertRuleFile(req *request.UpdateAlertRuleConfigRequest) error

	// AlertRule Options
	GetGroupList(ctx core.Context) response.GetGroupListResponse
	GetMetricPQL(ctx core.Context) (*response.GetMetricPQLResponse, error)

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
	chRepo   clickhouse.Repo
	promRepo prometheus.Repo
	k8sApi   kubernetes.Repo
	dbRepo   database.Repo
	difyRepo dify.DifyRepo

	enableInnerReceiver bool
	receivers           amreceiver.Receivers
}

func New(
	chRepo clickhouse.Repo,
	promRepo prometheus.Repo,
	k8sApi kubernetes.Repo,
	dbRepo database.Repo,
	difyRepo dify.DifyRepo,
	receivers amreceiver.Receivers,
) Service {

	cfg := config.Get().AMReceiver

	return &service{
		chRepo:   chRepo,
		promRepo: promRepo,
		k8sApi:   k8sApi,
		dbRepo:   dbRepo,
		difyRepo: difyRepo,

		enableInnerReceiver: cfg.Enabled,
		receivers:           receivers,
	}
}
