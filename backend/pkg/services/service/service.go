// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
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
	GetServiceEndpointRelation(ctx core.Context, req *request.GetServiceEndpointRelationRequest) (*response.GetServiceEndpointRelationResponse, error)
	// Get the upstream and downstream topology map
	GetServiceEndpointTopology(ctx core.Context, req *request.GetServiceEndpointTopologyRequest) (*response.GetServiceEndpointTopologyResponse, error)
	// Get the delay curve of the dependent service
	GetDescendantMetrics(ctx core.Context, req *request.GetDescendantMetricsRequest) ([]response.GetDescendantMetricsResponse, error)
	// Get the dependent node delay correlation.
	GetDescendantRelevance(ctx core.Context, req *request.GetDescendantRelevanceRequest) ([]response.GetDescendantRelevanceResponse, error)
	// Get Polaris metric analysis
	GetPolarisInfer(ctx core.Context, req *request.GetPolarisInferRequest) (*response.GetPolarisInferResponse, error)
	// Get error instance
	GetErrorInstance(ctx core.Context, req *request.GetErrorInstanceRequest) (*response.GetErrorInstanceResponse, error)
	// Get the error instance fault site log
	GetErrorInstanceLogs(ctx core.Context, req *request.GetErrorInstanceLogsRequest) ([]clickhouse.FaultLogResult, error)
	// Get log metrics
	GetLogMetrics(ctx core.Context, req *request.GetLogMetricsRequest) ([]*response.GetLogMetricsResponse, error)
	// Get Log fault field log
	GetLogLogs(ctx core.Context, req *request.GetLogLogsRequest) ([]clickhouse.FaultLogResult, error)
	// Get Trace related metrics
	GetTraceMetrics(ctx core.Context, req *request.GetTraceMetricsRequest) ([]*response.GetTraceMetricsResponse, error)
	// Get SQL related metrics
	GetSQLMetrics(ctx core.Context, req *request.GetSQLMetricsRequest) (*response.GetSQLMetricsResponse, error)
	// Get trace fault site log
	GetTraceLogs(ctx core.Context, req *request.GetTraceLogsRequest) ([]clickhouse.FaultLogResult, error)
	// Get the list of services
	GetServiceList(ctx core.Context, req *request.GetServiceListRequest) ([]string, error)
	// Get the list of service instances
	// New interface
	GetInstancesNew(ctx core.Context, startTime time.Time, endTime time.Time, step time.Duration, serviceName string, endPoint string) (res response.InstancesRes, err error)
	// Old interface
	GetInstances(ctx core.Context, startTime time.Time, endTime time.Time, step time.Duration, serviceName string, endPoint string) (res response.InstancesRes, err error)
	// Get the list of service instances
	// DEPRECATED
	GetServiceInstanceList(ctx core.Context, req *request.GetServiceInstanceListRequest) ([]string, error)
	// Get service instance details
	GetServiceInstanceInfoList(ctx core.Context, req *request.GetServiceInstanceListRequest) ([]prometheus.InstanceKey, error)
	// Get service instance drop-down list
	GetServiceInstanceOptions(ctx core.Context, req *request.GetServiceInstanceOptionsRequest) (map[string]*model.ServiceInstance, error)
	// Get the list of service Endpoint
	GetServiceEndPointList(ctx core.Context, req *request.GetServiceEndPointListRequest) ([]string, error)
	// Get the list of service portal Endpoint
	GetServiceEntryEndpoints(ctx core.Context, req *request.GetServiceEntryEndpointsRequest) ([]clickhouse.EntryNode, error)
	// CountK8sEvents get the number of K8s events
	CountK8sEvents(ctx core.Context, req *request.GetK8sEventsRequest) (*response.GetK8sEventsResponse, error)
	// GetAlertEvents get alarm events
	GetAlertEvents(ctx core.Context, req *request.GetAlertEventsRequest) (*response.GetAlertEventsResponse, error)

	// GetAlertEventsSample get sampled alarm events
	GetAlertEventsSample(ctx core.Context, req *request.GetAlertEventsSampleRequest) (*response.GetAlertEventsSampleResponse, error)
	GetServiceNamespaceList(ctx core.Context, req *request.GetServiceNamespaceListRequest) (response.GetServiceNamespaceListResponse, error)

	GetServiceREDCharts(ctx core.Context, req *request.GetServiceREDChartsRequest) (response.GetServiceREDChartsResponse, error)
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
