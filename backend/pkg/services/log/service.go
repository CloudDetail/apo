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
)

var _ Service = (*service)(nil)

type Service interface {
	// Get fault site paging log
	GetFaultLogPageList(req *request.GetFaultLogPageListRequest) (*response.GetFaultLogPageListResponse, error)

	GetFaultLogContent(req *request.GetFaultLogContentRequest) (*response.GetFaultLogContentResponse, error)

	InitParseLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error)

	DropLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error)

	UpdateLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error)

	GetLogTableInfo(req *request.LogTableInfoRequest) (*response.LogTableInfoResponse, error)

	// Query full logs
	QueryLog(req *request.LogQueryRequest) (*response.LogQueryResponse, error)

	QueryLogContext(req *request.LogQueryContextRequest) (*response.LogQueryContextResponse, error)
	// Log Trend Chart
	GetLogChart(req *request.LogQueryRequest) (*response.LogChartResponse, error)
	// Field Analysis
	GetLogIndex(req *request.LogIndexRequest) (*response.LogIndexResponse, error)

	GetServiceRoute(req *request.GetServiceRouteRequest) (*response.GetServiceRouteResponse, error)

	GetLogParseRule(req *request.QueryLogParseRequest) (*response.LogParseResponse, error)

	UpdateLogParseRule(req *request.UpdateLogParseRequest) (*response.LogParseResponse, error)

	AddLogParseRule(req *request.AddLogParseRequest) (*response.LogParseResponse, error)

	DeleteLogParseRule(req *request.DeleteLogParseRequest) (*response.LogParseResponse, error)

	OtherTable(req *request.OtherTableRequest) (*response.OtherTableResponse, error)

	OtherTableInfo(req *request.OtherTableInfoRequest) (*response.OtherTableInfoResponse, error)

	AddOtherTable(req *request.AddOtherTableRequest) (*response.AddOtherTableResponse, error)

	DeleteOtherTable(req *request.DeleteOtherTableRequest) (*response.DeleteOtherTableResponse, error)
}

type service struct {
	chRepo   clickhouse.Repo
	dbRepo   database.Repo
	k8sApi   kubernetes.Repo
	promRepo prometheus.Repo
}

func New(chRepo clickhouse.Repo, dbRepo database.Repo, k8sApi kubernetes.Repo, promRepo prometheus.Repo) Service {
	return &service{
		chRepo:   chRepo,
		dbRepo:   dbRepo,
		k8sApi:   k8sApi,
		promRepo: promRepo,
	}
}
