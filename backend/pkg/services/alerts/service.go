// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/receiver"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	// ========================告警检索========================
	AlertEventList(ctx core.Context, req *request.AlertEventSearchRequest) (*response.AlertEventSearchResponse, error)
	AlertDetail(ctx core.Context, req *request.GetAlertDetailRequest) (*response.GetAlertDetailResponse, error)

	AlertEventClassify(ctx core.Context, req *request.AlertEventClassifyRequest) (*response.AlertEventClassifyResponse, error)

	// ========================告警配置========================

	// InputAlertManager receive AlertManager alarm events
	// Deprecated: use alertinput.ProcessAlertEvents instead
	InputAlertManager(ctx core.Context, req *request.InputAlertManagerRequest) error
	ForwardToDingTalk(ctx core.Context, req *request.ForwardToDingTalkRequest, uuid string) error

	// GetAlertRuleFile get basic alarm rules
	GetAlertRuleFile(ctx core.Context, req *request.GetAlertRuleConfigRequest) (*response.GetAlertRuleFileResponse, error)
	// UpdateAlertRuleFile update basic alarm rules
	UpdateAlertRuleFile(ctx core.Context, req *request.UpdateAlertRuleConfigRequest) error

	// AlertRule Options
	GetGroupList(ctx core.Context) response.GetGroupListResponse
	GetMetricPQL(ctx core.Context) (*response.GetMetricPQLResponse, error)

	// AlertRule CRUD
	GetAlertRules(ctx core.Context, req *request.GetAlertRuleRequest) response.GetAlertRulesResponse
	UpdateAlertRule(ctx core.Context, req *request.UpdateAlertRuleRequest) error
	DeleteAlertRule(ctx core.Context, req *request.DeleteAlertRuleRequest) error
	AddAlertRule(ctx core.Context, req *request.AddAlertRuleRequest) error
	CheckAlertRule(ctx core.Context, req *request.CheckAlertRuleRequest) (response.CheckAlertRuleResponse, error)

	// AlertManager Receiver CRUD
	GetAMConfigReceivers(ctx core.Context, req *request.GetAlertManagerConfigReceverRequest) response.GetAlertManagerConfigReceiverResponse
	AddAMConfigReceiver(ctx core.Context, req *request.AddAlertManagerConfigReceiver) error
	UpdateAMConfigReceiver(ctx core.Context, req *request.UpdateAlertManagerConfigReceiver) error
	DeleteAMConfigReceiver(ctx core.Context, req *request.DeleteAlertManagerConfigReceiverRequest) error

	GetSlienceConfigByAlertID(ctx core.Context, alertID string) (*slienceconfig.AlertSlienceConfig, error)
	ListSlienceConfig(ctx core.Context) ([]slienceconfig.AlertSlienceConfig, error)
	SetSlienceConfigByAlertID(ctx core.Context, req *request.SetAlertSlienceConfigRequest) error
	RemoveSlienceConfigByAlertID(ctx core.Context, alertID string) error

	ManualResolveLatestAlertEventByAlertID(ctx core.Context, alertID string) error

	GetStaticFilterKeys(ctx core.Context) *response.AlertEventFiltersResponse
	GetAlertEventFilterLabelKeys(ctx core.Context, req *request.SearchAlertEventFilterValuesRequest) (*response.AlertEventFilterLabelKeysResponse, error)
	GetAlertEventFilterValues(ctx core.Context, req *request.SearchAlertEventFilterValuesRequest) (*request.AlertEventFilter, error)
}

type service struct {
	chRepo   clickhouse.Repo
	promRepo prometheus.Repo
	k8sApi   kubernetes.Repo
	dbRepo   database.Repo
	difyRepo dify.DifyRepo

	enableInnerReceiver bool
	receivers           receiver.Receivers
}

func New(
	chRepo clickhouse.Repo,
	promRepo prometheus.Repo,
	k8sApi kubernetes.Repo,
	dbRepo database.Repo,
	difyRepo dify.DifyRepo,
	receivers receiver.Receivers,
) Service {

	cfg := config.Get().AlertReceiver

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
