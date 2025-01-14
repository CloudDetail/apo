// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package deepflow

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/data"
	"github.com/CloudDetail/apo/backend/pkg/services/network"
	"go.uber.org/zap"
)

type Handler interface {
	// GetPodMap 查询 Pod 网络调用拓扑与调用指标
	// @Tags API.network
	// @Router /api/network/podmap [get]
	GetPodMap() core.HandlerFunc
	// GetSpanSegmentsMetrics 客户端对外调用Span网络耗时分段指标
	// @Tags API.network
	// @Router /api/network/segments [get]
	GetSpanSegmentsMetrics() core.HandlerFunc
}

type handler struct {
	logger         *zap.Logger
	networkService network.Service
	dataService    data.Service
}

func New(logger *zap.Logger, dbRepo database.Repo, chRepo clickhouse.Repo, promRepo prometheus.Repo, k8sRepo kubernetes.Repo) Handler {
	return &handler{
		logger:         logger,
		networkService: network.New(chRepo),
		dataService:    data.New(dbRepo, promRepo, k8sRepo),
	}
}
