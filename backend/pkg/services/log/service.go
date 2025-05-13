// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

var _ Service = (*service)(nil)

type Service interface {
	// Get fault site paging log
	GetFaultLogPageList(ctx_core core.Context, req *request.GetFaultLogPageListRequest) (*response.GetFaultLogPageListResponse, error)

	GetFaultLogContent(ctx_core core.Context, req *request.GetFaultLogContentRequest) (*response.GetFaultLogContentResponse, error)

	InitParseLogTable(ctx_core core.Context, req *request.LogTableRequest) (*response.LogTableResponse, error)

	DropLogTable(ctx_core core.Context, req *request.LogTableRequest) (*response.LogTableResponse, error)

	UpdateLogTable(ctx_core core.Context, req *request.LogTableRequest) (*response.LogTableResponse, error)

	GetLogTableInfo(ctx_core core.Context, req *request.LogTableInfoRequest) (*response.LogTableInfoResponse, error)

	// Query full logs
	QueryLog(ctx_core core.Context, req *request.LogQueryRequest) (*response.LogQueryResponse, error)

	QueryLogContext(ctx_core core.Context, req *request.LogQueryContextRequest) (*response.LogQueryContextResponse, error)
	// Log Trend Chart
	GetLogChart(ctx_core core.Context, req *request.LogQueryRequest) (*response.LogChartResponse, error)
	// Field Analysis
	GetLogIndex(ctx_core core.Context, req *request.LogIndexRequest) (*response.LogIndexResponse, error)

	GetServiceRoute(ctx_core core.Context, req *request.GetServiceRouteRequest) (*response.GetServiceRouteResponse, error)

	GetLogParseRule(ctx_core core.Context, req *request.QueryLogParseRequest) (*response.LogParseResponse, error)

	UpdateLogParseRule(ctx_core core.Context, req *request.UpdateLogParseRequest) (*response.LogParseResponse, error)

	AddLogParseRule(ctx_core core.Context, req *request.AddLogParseRequest) (*response.LogParseResponse, error)

	DeleteLogParseRule(ctx_core core.Context, req *request.DeleteLogParseRequest) (*response.LogParseResponse, error)

	OtherTable(ctx_core core.Context, req *request.OtherTableRequest) (*response.OtherTableResponse, error)

	OtherTableInfo(ctx_core core.Context, req *request.OtherTableInfoRequest) (*response.OtherTableInfoResponse, error)

	AddOtherTable(ctx_core core.Context, req *request.AddOtherTableRequest) (*response.AddOtherTableResponse, error)

	DeleteOtherTable(ctx_core core.Context, req *request.DeleteOtherTableRequest) (*response.DeleteOtherTableResponse, error)
}

type service struct {
	chRepo		clickhouse.Repo
	dbRepo		database.Repo
	k8sApi		kubernetes.Repo
	promRepo	prometheus.Repo
}

func New(chRepo clickhouse.Repo, dbRepo database.Repo, k8sApi kubernetes.Repo, promRepo prometheus.Repo) Service {
	return &service{
		chRepo:		chRepo,
		dbRepo:		dbRepo,
		k8sApi:		k8sApi,
		promRepo:	promRepo,
	}
}
