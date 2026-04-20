// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"log"
	"os"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dataplane"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	GetServices(ctx core.Context, req *request.QueryServicesRequest) *response.QueryServicesResponse
	ListServiceRedCharts(ctx core.Context, req *request.ListServiceRedChartsRequest) *response.ListServiceRedChartsResponse
	GetServiceRedCharts(ctx core.Context, req *request.QueryServiceRedChartsRequest) *response.QueryServiceRedChartsResponse
	GetServiceEndpointRedCharts(ctx core.Context, req *request.QueryServiceRedChartsRequest) *response.QueryServiceRedChartsResponse
	GetServiceEndpoints(ctx core.Context, req *request.QueryServiceEndpointsRequest) *response.QueryServiceEndpointsResponse
	GetServiceInstances(ctx core.Context, req *request.QueryServiceInstancesRequest) *response.QueryServiceInstancesResponse
	GetServiceName(ctx core.Context, req *request.QueryServiceNameRequest) *response.QueryServiceNameResponse
	GetServiceTopology(ctx core.Context, req *request.QueryTopologyRequest) *response.QueryTopologyResponse
	CheckDataSource(ctx core.Context, req *model.CheckDataSourceRequest) *model.CheckDataSourceResponse
	CleanRelateServiceByRuleId(ctx core.Context, req *request.CleanRuleServicesRequest) error
	CleanRelateServiceByDataSource(ctx core.Context, req *request.CleanDataSourceServicesRequest) error

	CreateCustomTopology(ctx core.Context, req *request.CreateCustomTopologyRequest) error
	DeleteCustomTopology(ctx core.Context, req *request.DeleteCustomTopologyRequest) error
	ListCustomTopology(ctx core.Context, req *request.ListCustomTopologyRequest) (*response.ListCustomTopologyResponse, error)
	CheckServiceNameRule(ctx core.Context, req *request.SetServiceNameRuleRequest) (*response.CheckServiceNameRuleResponse, error)
	SetServiceNameRule(ctx core.Context, req *request.SetServiceNameRuleRequest) error
	ListServiceNameRule(ctx core.Context, req *request.ListServiceNameRuleRequest) (*response.ListServiceNameRuleResponse, error)
	DeleteServiceNameRule(ctx core.Context, req *request.DeleteServiceNameRuleRequest) error
	ListAPPInfoLabelsKeys(ctx core.Context, req *request.QueryAPPInfoTagsRequest) (*response.QueryAPPInfoTagsResponse, error)
	ListAPPInfoLabelValues(ctx core.Context, req *request.QueryAPPInfoTagValuesRequest) (*response.QueryAPPInfoTagValuesResponse, error)

	DPService
	GetDatasources(ctx core.Context, req *request.GetDatasourcesRequest) (response.GetDatasourcesResponse, error)
	GetServiceWithInstances(ctx core.Context, req *request.GetServiceWithInstancesRequest) (response.GetServiceWithInstancesResponse, error)
	GetIntegratedCapabilities(ctx core.Context) (*response.GetIntegratedCapabilitiesResponse, error)
}

type service struct {
	chRepo        clickhouse.Repo
	promRepo      prometheus.Repo
	dbRepo        database.Repo
	dataplaneRepo dataplane.DataplaneRepo
}

func New(
	chRepo clickhouse.Repo,
	promRepo prometheus.Repo,
	dbRepo database.Repo,
	dataplaneRepo dataplane.DataplaneRepo,
) Service {
	service := &service{
		chRepo:        chRepo,
		promRepo:      promRepo,
		dbRepo:        dbRepo,
		dataplaneRepo: dataplaneRepo,
	}

	if os.Getenv("PRODUCT_MODE") == "apo" { // apo or syncause
		if err := setupAPODataPlane(core.EmptyCtx(), dbRepo); err != nil {
			log.Printf("init default apo data plane failed, err: %v", err)
		}
	}

	return service
}

func setupAPODataPlane(ctx core.Context, dbRepo database.Repo) error {
	// TODO check if apo has be deleted
	dps, err := dbRepo.ListDataPlane(ctx)
	if err != nil {
		// list dp failed, skip init apo
		return err
	}

	for _, v := range dps {
		if v.Typ == "apo" {
			return nil
		}
	}

	return dbRepo.CreateDataPlane(ctx, &integration.DataPlane{
		Name:   "APO",
		Note:   "Auto created by APO Backend",
		Typ:    "apo",
		Params: "{}",
		Capability: integration.JSONField[[]string]{
			Obj: []string{"RedMetric", "ServiceInstance", "Topology", "Log", "Trace", "Metrics"},
		},
		IsDeleted: false,
	})
}
