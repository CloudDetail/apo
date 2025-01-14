// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/data"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
	"go.uber.org/zap"
)

type Handler interface {
	// GetEndPointsData 获取endpoints服务列表
	// @Tags API.service
	// @Router /api/service/endpoints [get]
	GetEndPointsData() core.HandlerFunc
	// GetServicesAlert 获取Service的日志告警和状态灯信息
	// @Tags API.service
	// @Router /api/service/servicesAlert [get]
	GetServicesAlert() core.HandlerFunc
	// GetServiceMoreUrlList 获取服务的更多url列表
	// @Tags API.service
	// @Router /api/service/moreUrl [get]
	GetServiceMoreUrlList() core.HandlerFunc
	// GetThreshold 获取单个阈值配置信息
	// @Tags API.service
	// @Router /api/service/getThreshold [get]
	GetThreshold() core.HandlerFunc
	// SetThreshold 配置单个阈值配置信息
	// @Tags API.service
	// @Router /api/service/setThreshold [post]
	SetThreshold() core.HandlerFunc

	// GetRYGLight 获取红绿灯结果
	// @Tags API.service
	// @Router /api/service/ryglight [get]
	GetRYGLight() core.HandlerFunc

	// GetMonitorStatus 获取kuma监控的服务状态
	// @Tags API.service
	// @Router /api/service/monitor/status [get]
	GetMonitorStatus() core.HandlerFunc
}

type handler struct {
	logger          *zap.Logger
	promClient      prometheus.Repo
	serviceoverview serviceoverview.Service
	dataService     data.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, promClient prometheus.Repo, dbRepo database.Repo, k8sRepo kubernetes.Repo) Handler {
	return &handler{
		logger:          logger,
		promClient:      promClient,
		serviceoverview: serviceoverview.New(chRepo, dbRepo, promClient),
		dataService:     data.New(dbRepo, promClient, k8sRepo),
	}
}
