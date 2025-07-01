// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	GetServices(ctx core.Context, req *request.QueryServicesRequest) *response.QueryServicesResponse
	GetServiceRedCharts(ctx core.Context, req *request.QueryServiceRedChartsRequest) *response.QueryServiceRedChartsResponse
	GetServiceEndpointRedCharts(ctx core.Context, req *request.QueryServiceRedChartsRequest) *response.QueryServiceRedChartsResponse
	GetServiceEndpoints(ctx core.Context, req *request.QueryServiceEndpointsRequest) *response.QueryServiceEndpointsResponse
	GetServiceInstances(ctx core.Context, req *request.QueryServiceInstancesRequest) *response.QueryServiceInstancesResponse
	GetServiceName(ctx core.Context, req *request.QueryServiceNameRequest) *response.QueryServiceNameResponse
	GetServiceTopology(ctx core.Context, req *request.QueryTopologyRequest) *response.QueryTopologyResponse
}

type service struct {
	chRepo   clickhouse.Repo
	promRepo prometheus.Repo
	dbRepo   database.Repo
	
}

func New(
	chRepo clickhouse.Repo,
	promRepo prometheus.Repo,
	dbRepo database.Repo,
) Service {
	return &service{
		chRepo:   chRepo,
		promRepo: promRepo,
		dbRepo:   dbRepo,
	}
}
