// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type GetDescendantAnormalEventResponse struct {
	AlertEvents []clickhouse.PagedAlertEvent `json:"alertEvents"`
	Pagination  *model.Pagination            `json:"pagination"`
}

type GetDescendantDeltaAnormalEventResponse struct {
	// AnormalEvents []model.AnormalEvent `json:"anormalEvents"`
	// AnormalEvents map[int64][]model.AnormalEvent `json:"anormalEvents"`
	AnormalCount TempChartObject `json:"anormalCount"`

	OriginAnormalCounts  []DescendantAnormalCounts `json:"originAnormalCounts"`
	AlertTriggeredCounts []DescendantAnormalCounts `json:"alertTriggeredCounts"`
	FinalAnormalCounts   []DescendantAnormalCounts `json:"finalAnormalCounts"`

	OriginAnormalEvents []DescendantAnormalEventRecord `json:"originAnormalEvents"`
	DeltaAnormalEvents  []DescendantAnormalEventRecord `json:"deltaAnormalEvents"`
	FinalAnormalEvents  []DescendantAnormalEventRecord `json:"finalAnormalEvents"`
}

type DescendantAnormalCounts struct {
	model.EndpointKey

	AnormalCountsMap map[model.AnormalType]int64 `json:"anormalCounts"`
}

type DescendantAnormalEventRecord struct {
	model.EndpointKey

	AlertKey string `json:"alertKey"`

	AnormalType   model.AnormalType `json:"anormalType"`
	Timestamp     int64             `json:"timestamp"`
	AnormalObject string            `json:"anormalObject"`
	AnormalReason string            `json:"anormalReason"`
	AnormalMsg    string            `json:"anormalMsg"`

	AnormalStatus string `json:"anormalStatus"` // startFiring / updatedFiring / resolved
}

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

type GetPredefinedDetectExprResponse struct {
	PredefinedMetrics []model.PredefinedMetricExpr `json:"metrics"`
}

type GetDefectDetectExecListResponse struct {
	Pagination  model.Pagination       `json:"pagination"`
	MutatedRule []model.DetectMutation `json:"mutatedRule"`
}

type GetDefectDetectRuleListResponse struct {
	AlertRule  []model.DetectMutation `json:"alertRule"`
	Pagination model.Pagination       `json:"pagination"`
}
