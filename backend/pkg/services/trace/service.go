// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/jaeger"
	"go.uber.org/zap"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

var _ Service = (*service)(nil)

type Service interface {
	GetTraceFilters(ctx core.Context, req *request.GetTraceFiltersRequest) (*response.GetTraceFiltersResponse, error)
	GetTraceFilterValues(ctx core.Context, startTime, endTime time.Time, searchText string, filter request.SpanTraceFilter) (*response.GetTraceFilterValueResponse, error)
	GetTracePageList(ctx core.Context, req *request.GetTracePageListRequest) (*response.GetTracePageListResponse, error)
	GetOnOffCPU(ctx core.Context, req *request.GetOnOffCPURequest) (*response.GetOnOffCPUResponse, error)
	GetSingleTraceID(ctx core.Context, req *request.GetSingleTraceInfoRequest) (string, error)
	GetFlameGraphData(ctx core.Context, req *request.GetFlameDataRequest) (response.GetFlameDataResponse, error)
	GetProcessFlameGraphData(ctx core.Context, req *request.GetProcessFlameGraphRequest) (response.GetProcessFlameGraphResponse, error)
}

type service struct {
	chRepo     clickhouse.Repo
	jaegerRepo jaeger.JaegerRepo
	logger     *zap.Logger
}

func New(chRepo clickhouse.Repo, jaegerRepo jaeger.JaegerRepo, logger *zap.Logger) Service {
	return &service{
		chRepo:     chRepo,
		jaegerRepo: jaegerRepo,
		logger:     logger,
	}
}
