// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"go.uber.org/zap"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/services/alerts"
)

type Handler interface {
	// InputAlertManager get AlertManager alarm events
	// @Tags API.alerts
	// @Router /api/alerts/inputs/alertmanager [post]
	InputAlertManager() core.HandlerFunc

	// ForwardToDingTalk the received alarm is forwarded to the DingTalk
	// @Tags API.alerts
	// @Router /api/alerts/outputs/dingtalk/{uuid} [post]
	ForwardToDingTalk() core.HandlerFunc

	// GetAlertRuleFile get basic alarm rules
	// @Tags API.alerts
	// @Router /api/alerts/rules/file [get]
	GetAlertRuleFile() core.HandlerFunc

	// UpdateAlertRuleFile update basic alarm rules
	// @Tags API.alerts
	// @Router /api/alerts/rules/file [post]
	UpdateAlertRuleFile() core.HandlerFunc

	// GetAlertRules list alarm rules
	// @Tags API.alerts
	// @Router /api/alerts/rule/list [post]
	GetAlertRules() core.HandlerFunc

	// UpdateAlertRule update alarm rules
	// @Tags API.alerts
	// @Router /api/alerts/rule [post]
	UpdateAlertRule() core.HandlerFunc

	// DeleteAlertRule delete alarm rules
	// @Tags API.alerts
	// @Router /api/alerts/rule [delete]
	DeleteAlertRule() core.HandlerFunc

	// GetAlertManagerConfigReceiver list alarm notification objects
	// @Tags API.alerts
	// @Router /api/alerts/alertmanager/receiver/list [post]
	GetAlertManagerConfigReceiver() core.HandlerFunc

	// AddAlertManagerConfigReceiver new alarm notification object
	// @Tags API.alerts
	// @Router /api/alerts/alertmanager/receiver/add [post]
	AddAlertManagerConfigReceiver() core.HandlerFunc

	// UpdateAlertManagerConfigReceiver update alarm notification object
	// @Tags API.alerts
	// @Router /api/alerts/alertmanager/receiver [post]
	UpdateAlertManagerConfigReceiver() core.HandlerFunc

	// DeleteAlertManagerConfigReceiver delete alarm notification object
	// @Tags API.alerts
	// @Router /api/alerts/alertmanager/receiver [delete]
	DeleteAlertManagerConfigReceiver() core.HandlerFunc
	// GetGroupList get the corresponding interfaces of group and label
	// @Tags API.alerts
	// @Router /api/alerts/rule/groups [get]
	GetGroupList() core.HandlerFunc

	// GetMetricPQL get metrics and PQL in alarm rules
	// @Tags API.alerts
	// @Router /api/alerts/rule/metrics [get]
	GetMetricPQL() core.HandlerFunc

	// AddAlertRule new alarm rules
	// @Tags API.alerts
	// @Router /api/alerts/rule/add [post]
	AddAlertRule() core.HandlerFunc

	// CheckAlertRule check whether the alarm rule name is available
	// @Tags API.alerts
	// @Router /api/alerts/rule/available/file/group/alert [get]
	CheckAlertRule() core.HandlerFunc
}

type handler struct {
	logger       *zap.Logger
	alertService alerts.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, k8sRepo kubernetes.Repo, dbRepo database.Repo) Handler {
	return &handler{
		logger:       logger,
		alertService: alerts.New(chRepo, k8sRepo, dbRepo),
	}
}
