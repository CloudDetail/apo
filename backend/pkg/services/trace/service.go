// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/repository/jaeger"
	"go.uber.org/zap"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

var _ Service = (*service)(nil)

type Service interface {
	GetTraceFilters(ctx_core core.Context, startTime, endTime time.Time, needUpdate bool) (*response.GetTraceFiltersResponse, error)
	GetTraceFilterValues(ctx_core core.Context, startTime, endTime time.Time, searchText string, filter request.SpanTraceFilter) (*response.GetTraceFilterValueResponse, error)
	GetTracePageList(ctx_core core.Context, req *request.GetTracePageListRequest) (*response.GetTracePageListResponse, error)
	GetOnOffCPU(ctx_core core.Context, req *request.GetOnOffCPURequest) (*response.GetOnOffCPUResponse, error)
	GetSingleTraceID(ctx_core core.Context, req *request.GetSingleTraceInfoRequest) (string, error)
	GetFlameGraphData(ctx_core core.Context, req *request.GetFlameDataRequest) (response.GetFlameDataResponse, error)
	GetProcessFlameGraphData(ctx_core core.Context, req *request.GetProcessFlameGraphRequest) (response.GetProcessFlameGraphResponse, error)
}

type service struct {
	chRepo		clickhouse.Repo
	jaegerRepo	jaeger.JaegerRepo
	logger		*zap.Logger
}

func New(chRepo clickhouse.Repo, jaegerRepo jaeger.JaegerRepo, logger *zap.Logger) Service {
	return &service{
		chRepo:		chRepo,
		jaegerRepo:	jaegerRepo,
		logger:		logger,
	}
}
