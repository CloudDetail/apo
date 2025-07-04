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
	// GetEndPointsData get the list of endpoints services
	// @Tags API.service
	// @Router /api/service/endpoints [post]
	GetEndPointsData() core.HandlerFunc
	// GetServicesAlert get the log alarm and status light information of the Service
	// @Tags API.service
	// @Router /api/service/servicesAlert [post]
	GetServicesAlert() core.HandlerFunc
	// GetServiceMoreUrlList get more url list of services
	// @Tags API.service
	// @Router /api/service/moreUrl [get]
	GetServiceMoreUrlList() core.HandlerFunc
	// GetThreshold get the configuration information of a single threshold
	// @Tags API.service
	// @Router /api/service/getThreshold [get]
	GetThreshold() core.HandlerFunc
	// SetThreshold the configuration information of a single threshold.
	// @Tags API.service
	// @Router /api/service/setThreshold [post]
	SetThreshold() core.HandlerFunc

	// GetRYGLight get traffic light results
	// @Tags API.service
	// @Router /api/service/ryglight [get]
	GetRYGLight() core.HandlerFunc

	// GetMonitorStatus get the service status monitored by kuma
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
		serviceoverview: serviceoverview.New(logger, chRepo, dbRepo, promClient),
		dataService:     data.New(dbRepo, promClient, chRepo, k8sRepo),
	}
}
