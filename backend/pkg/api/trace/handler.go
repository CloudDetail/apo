// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/jaeger"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/data"
	"github.com/CloudDetail/apo/backend/pkg/services/trace"
	"go.uber.org/zap"
)

type Handler interface {
	// GetTraceFilters the available filters for querying the Trace list
	// @Tags API.trace
	// @Router /api/trace/pagelist/filters [get]
	GetTraceFilters() core.HandlerFunc
	// GetTraceFilterValue query the available values of the specified filter
	// @Tags API.trace
	// @Router /api/trace/pagelist/filter/value [post]
	GetTraceFilterValue() core.HandlerFunc
	// GetTracePageList to query the trace paging list
	// @Tags API.trace
	// @Router /api/trace/pagelist [post]
	GetTracePageList() core.HandlerFunc

	// GetOnOffCPU get span execution consumption
	// @Tags API.trace
	// @Router /api/trace/onoffcpu [get]
	GetOnOffCPU() core.HandlerFunc

	// GetSingleTraceInfo get single-link Trace details
	// @Tags API.trace
	// @Router /api/trace/info [get]
	GetSingleTraceInfo() core.HandlerFunc
	// GetFlameGraphData get the flame map data of the specified time period and specified conditions
	// @Tags API.trace
	// @Router /api/trace/flame [get]
	GetFlameGraphData() core.HandlerFunc
	// GetProcessFlameGraph capture and integrate process-level flame graph data
	// @Tags API.trace
	// @Router /api/trace/flame/process [get]
	GetProcessFlameGraph() core.HandlerFunc
}

type handler struct {
	logger       *zap.Logger
	traceService trace.Service
	dataService  data.Service
}

func New(logger *zap.Logger, dbRepo database.Repo, chRepo clickhouse.Repo, jaegerRepo jaeger.JaegerRepo, promRepo prometheus.Repo, k8sRepo kubernetes.Repo) Handler {
	return &handler{
		logger:       logger,
		traceService: trace.New(chRepo, dbRepo, jaegerRepo, logger),
		dataService:  data.New(dbRepo, promRepo, chRepo, k8sRepo),
	}
}
