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
	// @Router /api/service/relation [get]
	GetServiceEndpointRelation() core.HandlerFunc

	// GetServiceEndpointTopology get the upstream and downstream topology of a service
	// @Tags API.service
	// @Router /api/service/topology [get]
	GetServiceEndpointTopology() core.HandlerFunc

	// GetDescendantMetrics get the delay curve data of all downstream services
	// @Tags API.service
	// @Router /api/service/descendant/metrics [get]
	GetDescendantMetrics() core.HandlerFunc

	// GetDescendantRelevance get the dependent node delay correlation degree
	// @Tags API.service
	// @Router /api/service/descendant/relevance [get]
	GetDescendantRelevance() core.HandlerFunc

	// GetPolarisInfer access to Polaris metric analysis
	// @Tags API.service
	// @Router /api/service/polaris/infer [get]
	GetPolarisInfer() core.HandlerFunc

	// GetErrorInstance get error instance
	// @Tags API.service
	// @Router /api/service/error/instance [get]
	GetErrorInstance() core.HandlerFunc

	// GetErrorInstanceLogs get the error instance fault site log
	// @Tags API.service
	// @Router /api/service/errorinstance/logs [get]
	GetErrorInstanceLogs() core.HandlerFunc

	// GetLogMetrics get log metrics
	// @Tags API.service
	// @Router /api/service/log/metrics [get]
	GetLogMetrics() core.HandlerFunc

	// GetLogLogs get Log fault site log
	// @Tags API.service
	// @Router /api/service/log/logs [get]
	GetLogLogs() core.HandlerFunc

	// GetTraceMetrics get Trace related metrics
	// @Tags API.service
	// @Router /api/service/trace/metrics [get]
	GetTraceMetrics() core.HandlerFunc

	// GetTraceLogs get trace fault site log
	// @Tags API.service
	// @Router /api/service/trace/logs [get]
	GetTraceLogs() core.HandlerFunc

	// GetServiceList get the list of services
	// @Tags API.service
	// @Router /api/service/list [get]
	GetServiceList() core.HandlerFunc

	// GetServiceInstance get the list of service instances
	// @Tags API.service
	// @Router /api/service/instance [get]
	GetServiceInstance() core.HandlerFunc

	// GetServiceInstanceList get more url list of services
	// @Tags API.service
	// @DEPRECATED
	// @Router /api/service/instances/list [get]
	GetServiceInstanceList() core.HandlerFunc

	// GetServiceInstanceList get the details of the service instance.
	// @Tags API.service
	// @Router /api/service/instanceinfo/list [get]
	GetServiceInstanceInfoList() core.HandlerFunc

	// GetServiceInstanceOptions get the drop-down list of service instances
	// @Tags API.service
	// @Router /api/service/instance/options [get]
	GetServiceInstanceOptions() core.HandlerFunc

	// GetServiceEndPointList get the list of service EndPoint
	// @Tags API.service
	// @Router /api/service/endpoint/list [get]
	GetServiceEndPointList() core.HandlerFunc

	// CountK8sEvents get the number of K8s events
	// @Tags API.service
	// @Router /api/service/k8s/events/count [get]
	CountK8sEvents() core.HandlerFunc

	// GetAlertEventsSample get sampling alarm events
	// @Tags API.service
	// @Router /api/service/alert/sample/events [get]
	GetAlertEventsSample() core.HandlerFunc

	// GetAlertEvents get alarm events
	// @Tags API.service
	// @Router /api/service/alert/events [get]
	GetAlertEvents() core.HandlerFunc

	// Get SQL metrics GetSQLMetrics
	// @Tags API.service
	// @Router /api/service/sql/metrics [get]
	GetSQLMetrics() core.HandlerFunc

	// GetServiceEntryEndpoints get the list of service portal Endpoint
	// @Tags API.service
	// @Router /api/service/entry/endpoints [get]
	GetServiceEntryEndpoints() core.HandlerFunc

	// GetNamespaceList Get monitored namespaces.
	// @Tags API.service
	// @Router /api/service/namespace/list [get]
	GetNamespaceList() core.HandlerFunc
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
		serviceInfoService:     service.New(chRepo, promRepo, polRepo, dbRepo),
		serviceoverviewService: serviceoverview.New(chRepo, dbRepo, promRepo),
		dataService:            data.New(dbRepo, promRepo, k8sRepo),
	}
}
