// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	alertinput "github.com/CloudDetail/apo/backend/pkg/services/integration/alert"
	"go.uber.org/zap"
)

type Handler interface {
	// JsonHandler Receive data from a specific data source based on a JSON structure
	// @Tags API.alertinput
	// @Router /api/alertinput/event/json [post]
	JsonHandler() core.HandlerFunc

	// SourceHandler Receive data based on alarm source configuration
	// @Tags API.alertinput
	// @Router /api/alertinput/event/source [post]
	SourceHandler() core.HandlerFunc

	// CreateAlertSource Create Alarm Source
	// @Tags API.alertinput
	// @Router /api/alertinput/source/create [post]
	CreateAlertSource() core.HandlerFunc

	// GetAlertSource Obtain alarm source information
	// @Tags API.alertinput
	// @Router /api/alertinput/source/get [post]
	GetAlertSource() core.HandlerFunc

	// UpdateAlertSource Update alarm source
	// @Tags API.alertinput
	// @Router /api/alertinput/source/update [post]
	UpdateAlertSource() core.HandlerFunc

	// DeleteAlertSource Delete Alarm Source
	// @Tags API.alertinput
	// @Router /api/alertinput/source/delete [post]
	DeleteAlertSource() core.HandlerFunc

	// ListAlertSource List alarm sources
	// @Tags API.alertinput
	// @Router /api/alertinput/source/list [get]
	ListAlertSource() core.HandlerFunc

	// UpdateAlertSourceEnrichRule Update alarm source enhanced configuration
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/update [post]
	UpdateAlertSourceEnrichRule() core.HandlerFunc

	// GetAlertSourceEnrichRule Obtain alarm source enhancement configuration
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/get [get]
	GetAlertSourceEnrichRule() core.HandlerFunc

	// ListTargetTags Obtain predefined labels for association
	// @Tags API.alertinput
	// @Router /api/alertinput/enrich/tags/list [get]
	ListTargetTags() core.HandlerFunc

	// ListCluster ListCluster
	// @Tags API.alertinput
	// @Router /api/alertinput/cluster/list [get]
	ListCluster() core.HandlerFunc

	// CreateCluster CreateCluster
	// @Tags API.alertinput
	// @Router /api/alertinput/cluster/create [post]
	CreateCluster() core.HandlerFunc

	// UpdateCluster UpdateCluster
	// @Tags API.alertinput
	// @Router /api/alertinput/cluster/update [post]
	UpdateCluster() core.HandlerFunc

	// DeleteCluster DeleteCluster
	// @Tags API.alertinput
	// @Router /api/alertinput/cluster/delete [post]
	DeleteCluster() core.HandlerFunc

	// CreateSchema CreateSchema
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/create [post]
	CreateSchema() core.HandlerFunc

	// DeleteSchema DeleteSchema
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/delete [get]
	DeleteSchema() core.HandlerFunc

	// ListSchema ListSchema
	// @Tags API.ListSchema
	// @Router /api/alertinput/schema/list [get]
	ListSchema() core.HandlerFunc

	// GetSchemaColumns GetSchemaColumns
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/column/get [get]
	GetSchemaColumns() core.HandlerFunc

	// UpdateSchemaData UpdateSchemaData
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/data/update [post]
	UpdateSchemaData() core.HandlerFunc

	// GetSchemaData core.HandlerFunc
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/data/get [get]
	GetSchemaData() core.HandlerFunc

	// CheckSchemaIsUsed CheckSchemaIsUsed
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/used/check [get]
	CheckSchemaIsUsed() core.HandlerFunc

	// GetDefaultAlertEnrichRule GetDefaultAlertEnrichRule
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/default [get]
	GetDefaultAlertEnrichRule() core.HandlerFunc

	// ClearDefaultAlertEnrichRule ClearDefaultAlertEnrichRule
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/default/clear [get]
	ClearDefaultAlertEnrichRule() core.HandlerFunc

	// SetDefaultAlertEnrichRule SetDefaultAlertEnrichRule
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/default/set [post]
	SetDefaultAlertEnrichRule() core.HandlerFunc

	// ListSchemaWithColumns ListSchemaWithColumns
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/listwithcolumns [get]
	ListSchemaWithColumns() core.HandlerFunc

	// SetIncidentTempBySource SetIncidentTempBySource
	// @Tags API.alertinput
	// @Router /api/alertinput/incident/temp/set [post]
	SetIncidentTempBySource() core.HandlerFunc

	// GetIncidentTempBySource GetIncidentTempBySource
	// @Tags API.alertinput
	// @Router /api/alertinput/incident/temp/get [get]
	GetIncidentTempBySource() core.HandlerFunc

	// ClearIncidentTempBySource ClearIncidentTempBySource
	// @Tags API.alertinput
	// @Router /api/alertinput/incident/temp/clear [get]
	ClearIncidentTempBySource() core.HandlerFunc
}

type handler struct {
	logger       *zap.Logger
	inputService alertinput.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, promRepo prometheus.Repo, dbRepo database.Repo, difyRepo dify.DifyRepo) Handler {
	return &handler{
		logger:       logger,
		inputService: alertinput.New(promRepo, dbRepo, chRepo, difyRepo),
	}
}
