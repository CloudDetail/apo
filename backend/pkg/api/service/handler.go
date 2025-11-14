// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/services/data"
	"go.uber.org/zap"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/service"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

type Handler interface {
	// GetServiceEndpointRelation get the call relationship between the upstream and downstream services.
	// @Tags API.service
	// @Router /api/service/relation [post]
	GetServiceEndpointRelation() core.HandlerFunc

	// GetServiceEndpointTopology get the upstream and downstream topology of a service
	// @Tags API.service
	// @Router /api/service/topology [post]
	GetServiceEndpointTopology() core.HandlerFunc

	// GetDescendantMetrics get the delay curve data of all downstream services
	// @Tags API.service
	// @Router /api/service/descendant/metrics [post]
	GetDescendantMetrics() core.HandlerFunc

	// GetDescendantRelevance get the dependent node delay correlation degree
	// @Tags API.service
	// @Router /api/service/descendant/relevance [post]
	GetDescendantRelevance() core.HandlerFunc

	// GetPolarisInfer access to Polaris metric analysis
	// @Tags API.service
	// @Router /api/service/polaris/infer [post]
	GetPolarisInfer() core.HandlerFunc

	// GetErrorInstance get error instance
	// @Tags API.service
	// @Router /api/service/error/instance [post]
	GetErrorInstance() core.HandlerFunc

	// GetErrorInstanceLogs get the error instance fault site log
	// @Tags API.service
	// @Router /api/service/errorinstance/logs [post]
	GetErrorInstanceLogs() core.HandlerFunc

	// GetLogMetrics get log metrics
	// @Tags API.service
	// @Router /api/service/log/metrics [post]
	GetLogMetrics() core.HandlerFunc

	// GetLogLogs get Log fault site log
	// @Tags API.service
	// @Router /api/service/log/logs [post]
	GetLogLogs() core.HandlerFunc

	// GetTraceMetrics get Trace related metrics
	// @Tags API.service
	// @Router /api/service/trace/metrics [get]
	GetTraceMetrics() core.HandlerFunc

	// GetTraceLogs get trace fault site log
	// @Tags API.service
	// @Router /api/service/trace/logs [post]
	GetTraceLogs() core.HandlerFunc

	// GetServiceList get the list of services
	// @Tags API.service
	// @Router /api/service/list [post]
	GetServiceList() core.HandlerFunc

	// GetServiceInstance get the list of service instances
	// @Tags API.service
	// @Router /api/service/instance [post]
	GetServiceInstance() core.HandlerFunc

	// GetServiceInstanceList get more url list of services
	// @Tags API.service
	// @DEPRECATED
	// @Router /api/service/instances/list [post]
	GetServiceInstanceList() core.HandlerFunc

	// GetServiceInstanceList get the details of the service instance.
	// @Tags API.service
	// @Router /api/service/instanceinfo/list [post]
	GetServiceInstanceInfoList() core.HandlerFunc

	// GetServiceInstanceOptions get the drop-down list of service instances
	// @Tags API.service
	// @Router /api/service/instance/options [post]
	GetServiceInstanceOptions() core.HandlerFunc

	// GetServiceEndPointList get the list of service EndPoint
	// @Tags API.service
	// @Router /api/service/endpoint/list [post]
	GetServiceEndPointList() core.HandlerFunc

	// CountK8sEvents get the number of K8s events
	// @Tags API.service
	// @Router /api/service/k8s/events/count [post]
	CountK8sEvents() core.HandlerFunc

	// GetAlertEventsSample get sampling alarm events
	// @Tags API.service
	// @Router /api/service/alert/sample/events [post]
	GetAlertEventsSample() core.HandlerFunc

	// GetAlertEvents get alarm events
	// @Tags API.service
	// @Router /api/service/alert/events [post]
	GetAlertEvents() core.HandlerFunc

	// Get SQL metrics GetSQLMetrics
	// @Tags API.service
	// @Router /api/service/sql/metrics [post]
	GetSQLMetrics() core.HandlerFunc

	// GetServiceEntryEndpoints get the list of service portal Endpoint
	// @Tags API.service
	// @Router /api/service/entry/endpoints [post]
	GetServiceEntryEndpoints() core.HandlerFunc

	// GetNamespaceList Get monitored namespaces.
	// @Tags API.service
	// @Router /api/service/namespace/list [post]
	GetNamespaceList() core.HandlerFunc

	// GetServiceREDCharts Get services' red charts.
	// @Tags API.service
	// @Router /api/service/redcharts [post]
	GetServiceREDCharts() core.HandlerFunc
}

type handler struct {
	logger                 *zap.Logger
	serviceInfoService     service.Service
	serviceoverviewService serviceoverview.Service
	dataService            data.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, promRepo prometheus.Repo, polRepo polarisanalyzer.Repo, dbRepo database.Repo, k8sRepo kubernetes.Repo) Handler {
	return &handler{
		logger:                 logger,
		serviceInfoService:     service.New(chRepo, promRepo, polRepo, dbRepo, logger),
		serviceoverviewService: serviceoverview.New(logger, chRepo, dbRepo, promRepo),
		dataService:            data.New(dbRepo, promRepo, chRepo, k8sRepo),
	}
}
