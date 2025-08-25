// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/provider"
)

type GetAlertRuleFileResponse struct {
	AlertRules map[string]string `json:"alertRules"`
}

type GetAlertRulesResponse struct {
	AlertRules []*request.AlertRule `json:"alertRules"`

	Pagination *model.Pagination `json:"pagination"`
}

type GetAlertManagerConfigReceiverResponse struct {
	AMConfigReceivers []amconfig.Receiver `json:"amConfigReceivers"`

	Pagination *model.Pagination `json:"pagination"`
}

type GetGroupListResponse struct {
	GroupsLabel map[string]string `json:"groupsLabel"`
}

type GetMetricPQLResponse struct {
	AlertMetricsData []database.AlertMetricsData `json:"alertMetricsData"`
}

type CheckAlertRuleResponse struct {
	Available bool `json:"available"`
}

type GetAlertProviderParamsSpecResponse struct {
	ParamSpec *provider.ParamSpec `json:"paramSpec"`

	WithPullOptions bool `json:"withPullOptions"`
}
