// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	// Get the relationship between upstream and downstream calls
	GetServiceEndpointRelation(req *request.GetServiceEndpointRelationRequest) (*response.GetServiceEndpointRelationResponse, error)
	// Get the upstream and downstream topology map
	GetServiceEndpointTopology(req *request.GetServiceEndpointTopologyRequest) (*response.GetServiceEndpointTopologyResponse, error)
	// Get the delay curve of the dependent service
	GetDescendantMetrics(req *request.GetDescendantMetricsRequest) ([]response.GetDescendantMetricsResponse, error)
	// Get the dependent node delay correlation.
	GetDescendantRelevance(ctx core.Context, req *request.GetDescendantRelevanceRequest) ([]response.GetDescendantRelevanceResponse, error)
	// Get Polaris metric analysis
	GetPolarisInfer(req *request.GetPolarisInferRequest) (*response.GetPolarisInferResponse, error)
	// Get error instance
	GetErrorInstance(req *request.GetErrorInstanceRequest) (*response.GetErrorInstanceResponse, error)
	// Get the error instance fault site log
	GetErrorInstanceLogs(req *request.GetErrorInstanceLogsRequest) ([]clickhouse.FaultLogResult, error)
	// Get log metrics
	GetLogMetrics(req *request.GetLogMetricsRequest) ([]*response.GetLogMetricsResponse, error)
	// Get Log fault field log
	GetLogLogs(req *request.GetLogLogsRequest) ([]clickhouse.FaultLogResult, error)
	// Get Trace related metrics
	GetTraceMetrics(req *request.GetTraceMetricsRequest) ([]*response.GetTraceMetricsResponse, error)
	// Get SQL related metrics
	GetSQLMetrics(req *request.GetSQLMetricsRequest) (*response.GetSQLMetricsResponse, error)
	// Get trace fault site log
	GetTraceLogs(req *request.GetTraceLogsRequest) ([]clickhouse.FaultLogResult, error)
	// Get the list of services
	GetServiceList(req *request.GetServiceListRequest) ([]string, error)
	// Get the list of service instances
	// New interface
	GetInstancesNew(ctx core.Context, startTime time.Time, endTime time.Time, step time.Duration, serviceName string, endPoint string) (res response.InstancesRes, err error)
	// Old interface
	GetInstances(ctx core.Context, startTime time.Time, endTime time.Time, step time.Duration, serviceName string, endPoint string) (res response.InstancesRes, err error)
	// Get the list of service instances
	// DEPRECATED
	GetServiceInstanceList(req *request.GetServiceInstanceListRequest) ([]string, error)
	// Get service instance details
	GetServiceInstanceInfoList(req *request.GetServiceInstanceListRequest) ([]prometheus.InstanceKey, error)
	// Get service instance drop-down list
	GetServiceInstanceOptions(req *request.GetServiceInstanceOptionsRequest) (map[string]*model.ServiceInstance, error)
	// Get the list of service Endpoint
	GetServiceEndPointList(req *request.GetServiceEndPointListRequest) ([]string, error)
	// Get the list of service portal Endpoint
	GetServiceEntryEndpoints(req *request.GetServiceEntryEndpointsRequest) ([]clickhouse.EntryNode, error)
	// CountK8sEvents get the number of K8s events
	CountK8sEvents(req *request.GetK8sEventsRequest) (*response.GetK8sEventsResponse, error)
	// GetAlertEvents get alarm events
	GetAlertEvents(req *request.GetAlertEventsRequest) (*response.GetAlertEventsResponse, error)

	// GetAlertEventsSample get sampled alarm events
	GetAlertEventsSample(req *request.GetAlertEventsSampleRequest) (*response.GetAlertEventsSampleResponse, error)
	GetServiceNamespaceList(req *request.GetServiceNamespaceListRequest) (response.GetServiceNamespaceListResponse, error)

	GetServiceREDCharts(req *request.GetServiceREDChartsRequest) (response.GetServiceREDChartsResponse, error)
}

type service struct {
	chRepo   clickhouse.Repo
	promRepo prometheus.Repo
	polRepo  polarisanalyzer.Repo
	dbRepo   database.Repo
}

func New(chRepo clickhouse.Repo, promRepo prometheus.Repo, polRepo polarisanalyzer.Repo, dbRepo database.Repo) Service {
	return &service{
		chRepo:   chRepo,
		promRepo: promRepo,
		polRepo:  polRepo,
		dbRepo:   dbRepo,
	}
}
