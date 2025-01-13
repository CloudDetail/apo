// 接收输入的告警事件信息
// 使用http接收告警事件时,走原有的handler结构

// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	alertinput "github.com/CloudDetail/apo/backend/pkg/services/input/alert"
	"go.uber.org/zap"
)

type Handler interface {
	// JsonHandler 基于JSON结构接收来自特定数据源的数据
	// @Tags API.alertinput
	// @Router /api/alertinput/event/json/:sourceType/:sourceName [post]
	JsonHandler() core.HandlerFunc

	// SourceHandler 基于告警源配置接收数据
	// @Tags API.alertinput
	// @Router /api/alertinput/event/source/:sourceID [post]
	SourceHandler() core.HandlerFunc

	// CreateAlertSource 创建告警源
	// @Tags API.alertinput
	// @Router /api/alertinput/source/create [post]
	CreateAlertSource() core.HandlerFunc

	// GetAlertSource 获取告警源信息
	// @Tags API.alertinput
	// @Router /api/alertinput/source/get [post]
	GetAlertSource() core.HandlerFunc

	// UpdateAlertSource 更新告警源
	// @Tags API.alertinput
	// @Router /api/alertinput/source/update [post]
	UpdateAlertSource() core.HandlerFunc

	// DeleteAlertSource 删除告警源
	// @Tags API.alertinput
	// @Router /api/alertinput/source/delete [post]
	DeleteAlertSource() core.HandlerFunc

	// ListAlertSource 列出告警源
	// @Tags API.alertinput
	// @Router /api/alertinput/source/list [get]
	ListAlertSource() core.HandlerFunc

	// UpdateAlertSourceEnrichRule 更新告警源增强配置
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/update [post]
	UpdateAlertSourceEnrichRule() core.HandlerFunc

	// GetAlertSourceEnrichRule 获取告警源增强配置
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/get [get]
	GetAlertSourceEnrichRule() core.HandlerFunc

	// ListTargetTags 获取预先定义的关联用标签
	// @Tags API.alertinput
	// @Router /api/alertinput/enrich/tags/list [get]
	ListTargetTags() core.HandlerFunc

	// ListCluster 列出集群
	// @Tags API.alertinput
	// @Router /api/alertinput/cluster/list [get]
	ListCluster() core.HandlerFunc

	// CreateCluster 创建集群
	// @Tags API.alertinput
	// @Router /api/alertinput/cluster/create [post]
	CreateCluster() core.HandlerFunc

	// UpdateCluster 更新集群
	// @Tags API.alertinput
	// @Router /api/alertinput/cluster/update [post]
	UpdateCluster() core.HandlerFunc

	// DeleteCluster 删除集群
	// @Tags API.alertinput
	// @Router /api/alertinput/cluster/delete [post]
	DeleteCluster() core.HandlerFunc

	// CreateSchema 创建映射结构
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/create [post]
	CreateSchema() core.HandlerFunc

	// DeleteSchema 删除映射结构
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/delete [get]
	DeleteSchema() core.HandlerFunc

	// ListSchema 列出映射结构
	// @Tags API.ListSchema
	// @Router /api/alertinput/schema/list [get]
	ListSchema() core.HandlerFunc

	// GetSchemaColumns 获取映射结构中的列信息
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/column/get [get]
	GetSchemaColumns() core.HandlerFunc

	// UpdateSchemaData 更新映射结构中的数据
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/data/update [post]
	UpdateSchemaData() core.HandlerFunc

	// GetSchemaData core.HandlerFunc
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/data/get [get]
	GetSchemaData() core.HandlerFunc

	// CheckSchemaIsUsed 检查映射结构是否被使用
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/used/check [get]
	CheckSchemaIsUsed() core.HandlerFunc

	// GetDefaultAlertEnrichRule 获取默认的告警丰富规则
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/default [get]
	GetDefaultAlertEnrichRule() core.HandlerFunc

	// ClearDefaultAlertEnrichRule 清除默认的告警丰富规则
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/default/clear [get]
	ClearDefaultAlertEnrichRule() core.HandlerFunc

	// SetDefaultAlertEnrichRule 设置默认的告警丰富规则
	// @Tags API.alertinput
	// @Router /api/alertinput/source/enrich/default/set [post]
	SetDefaultAlertEnrichRule() core.HandlerFunc

	// ListSchemaWithColumns 列出映射表及结构
	// @Tags API.alertinput
	// @Router /api/alertinput/schema/listwithcolumns [get]
	ListSchemaWithColumns() core.HandlerFunc
}

type handler struct {
	logger       *zap.Logger
	inputService alertinput.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, promRepo prometheus.Repo, dbRepo database.Repo) Handler {
	return &handler{
		logger:       logger,
		inputService: alertinput.New(promRepo, dbRepo, chRepo),
	}
}
