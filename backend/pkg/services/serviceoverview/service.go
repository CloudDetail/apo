// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"go.uber.org/zap"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	GetServiceMoreUrl(ctx core.Context, startTime time.Time, endTime time.Time, step time.Duration, serviceNames string, sortRule request.SortType) (res []response.ServiceDetail, err error)
	GetThreshold(ctx core.Context, level string, serviceName string, endPoint string) (res response.GetThresholdResponse, err error)
	SetThreshold(ctx core.Context, level string, serviceName string, endPoint string, latency float64, errorRate float64, tps float64, log float64) (res response.SetThresholdResponse, err error)
	GetServicesAlert(ctx core.Context, startTime time.Time, endTime time.Time, step time.Duration, serviceNames []string, returnData []string) (res []response.ServiceAlertRes, err error)
	GetServicesEndPointData(ctx core.Context, req *request.GetEndPointsDataRequest) (res []response.ServiceEndPointsRes, err error)

	// TODO move to prometheus package and avoid to repeated again
	GetServicesEndpointDataWithChart(ctx core.Context, startTime time.Time, endTime time.Time, step time.Duration, filter EndpointsFilter, sortRule request.SortType) (res []response.ServiceEndPointsRes, err error)

	GetServicesRYGLightStatus(ctx core.Context, startTime time.Time, endTime time.Time, filter EndpointsFilter) (response.ServiceRYGLightRes, error)
	GetMonitorStatus(ctx core.Context, startTime time.Time, endTime time.Time) (response.GetMonitorStatusResponse, error)

	GetAlertRelatedEntryData(ctx core.Context, startTime, endTime time.Time, namespaces []string, entry []response.AlertRelatedEntry) (res []response.AlertRelatedEntry, err error)
}

type service struct {
	logger *zap.Logger

	dbRepo   database.Repo
	promRepo prometheus.Repo
	chRepo   clickhouse.Repo
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, dbRepo database.Repo, promRepo prometheus.Repo) Service {
	return &service{
		logger:   logger,
		dbRepo:   dbRepo,
		promRepo: promRepo,
		chRepo:   chRepo,
	}
}

type EndpointsFilter struct {
	ContainsSvcName      string   // SvcName, containing matches
	ContainsEndpointName string   // EndpointName, containing matches
	Namespace            string   // Namespace, exact match
	ServiceName          string   // Specify the service name, exact match
	MultiService         []string // multiple service names, exact match
	MultiNamespace       []string // multiple namespace, exact match
	MultiEndpoint        []string // multiple service endpoints, exact match

	ClusterIDs []string
}
