// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/data"
	"github.com/CloudDetail/apo/backend/pkg/services/log"
	"go.uber.org/zap"
)

type Handler interface {
	// GetFaultLogPageList get the fault site paging log
	// @Tags API.log
	// @Router /api/log/fault/pagelist [post]
	GetFaultLogPageList() core.HandlerFunc

	// GetFaultLogContent get the contents of the fault site log
	// @Tags API.log
	// @Router /api/log/fault/content [post]
	GetFaultLogContent() core.HandlerFunc

	// QueryLog query full logs
	// @Tags API.log
	// @Router /api/log/query [post]
	QueryLog() core.HandlerFunc

	// GetLogChart get the log trend chart
	// @Tags API.log
	// @Router /api/log/chart [post]
	GetLogChart() core.HandlerFunc

	// GetLogIndex analysis field index
	// @Tags API.log
	// @Router /api/log/index [post]
	GetLogIndex() core.HandlerFunc

	// GetLogTableInfo get log table information
	// @Tags API.log
	// @Router /api/log/table get
	GetLogTableInfo() core.HandlerFunc

	// GetServiceRoute get the application log corresponding to the service
	// @Tags API.log
	// @Router /api/log/rule/service get
	GetServiceRoute() core.HandlerFunc

	// GetLogParseRule get log table parsing rules
	// @Tags API.log
	// @Router /api/log/rule/get [get]
	GetLogParseRule() core.HandlerFunc

	// UpdateLogParseRule update log table parsing rules
	// @Tags API.log
	// @Router /api/log/rule/update [post]
	UpdateLogParseRule() core.HandlerFunc

	// AddLogParseRule new log table parsing rules
	// @Tags API.log
	// @Router /api/log/rule/add [post]
	AddLogParseRule() core.HandlerFunc

	// DeleteLogParseRule delete log table parsing rules
	// @Tags API.log
	// @Router /api/log/rule/delete [delete]
	DeleteLogParseRule() core.HandlerFunc

	// OtherTable get the external log table
	// @Tags API.log
	// @Router /api/log/other get
	OtherTable() core.HandlerFunc

	// OtherTableInfo get external log table information
	// @Tags API.log
	// @Router /api/log/other/table [get]
	OtherTableInfo() core.HandlerFunc

	// AddOtherTable add external log table
	// @Tags API.log
	// @Router /api/log/other/add [post]
	AddOtherTable() core.HandlerFunc

	// DeleteOtherTable remove external log table
	// @Tags API.log
	// @Router /api/log/other/delete [delete]
	DeleteOtherTable() core.HandlerFunc

	// QueryLogContext get the log context
	// @Tags API.log
	// @Router /api/log/context [post]
	QueryLogContext() core.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	logService  log.Service
	dataService data.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, dbRepo database.Repo, k8sApi kubernetes.Repo, promRepo prometheus.Repo) Handler {
	logservice := log.New(chRepo, dbRepo, k8sApi, promRepo)
	req := &request.LogTableRequest{}
	req.FillerValue()
	// TODO ctx
	_, err := logservice.InitParseLogTable(nil, req)
	if err != nil {
		logger.Error("create default log table failed", zap.Error(err))
	}
	return &handler{
		logger:      logger,
		logService:  logservice,
		dataService: data.New(dbRepo, promRepo, k8sApi),
	}
}
