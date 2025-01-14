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
	// GetFaultLogPageList 获取故障现场分页日志
	// @Tags API.log
	// @Router /api/log/fault/pagelist [post]
	GetFaultLogPageList() core.HandlerFunc

	// GetFaultLogContent 获取故障现场日志内容
	// @Tags API.log
	// @Router /api/log/fault/content [post]
	GetFaultLogContent() core.HandlerFunc

	// QueryLog 查询全量日志
	// @Tags API.log
	// @Router /api/log/query [post]
	QueryLog() core.HandlerFunc

	// GetLogChart 获取日志趋势图
	// @Tags API.log
	// @Router /api/log/chart [post]
	GetLogChart() core.HandlerFunc

	// GetLogIndex 分析字段索引
	// @Tags API.log
	// @Router /api/log/index [post]
	GetLogIndex() core.HandlerFunc

	// GetLogTableInfo 获取日志表信息
	// @Tags API.log
	// @Router /api/log/table get
	GetLogTableInfo() core.HandlerFunc

	// GetServiceRoute 获取服务对应的应用日志
	// @Tags API.log
	// @Router /api/log/rule/service get
	GetServiceRoute() core.HandlerFunc

	// GetLogParseRule 获取日志表解析规则
	// @Tags API.log
	// @Router /api/log/rule/get [get]
	GetLogParseRule() core.HandlerFunc

	// UpdateLogParseRule 更新日志表解析规则
	// @Tags API.log
	// @Router /api/log/rule/update [post]
	UpdateLogParseRule() core.HandlerFunc

	// AddLogParseRule 新增日志表解析规则
	// @Tags API.log
	// @Router /api/log/rule/add [post]
	AddLogParseRule() core.HandlerFunc

	// DeleteLogParseRule 删除日志表解析规则
	// @Tags API.log
	// @Router /api/log/rule/delete [delete]
	DeleteLogParseRule() core.HandlerFunc

	// OtherTable 获取外部日志表
	// @Tags API.log
	// @Router /api/log/other get
	OtherTable() core.HandlerFunc

	// OtherTableInfo 获取外部日志表信息
	// @Tags API.log
	// @Router /api/log/other/table [get]
	OtherTableInfo() core.HandlerFunc

	// AddOtherTable 添加外部日志表
	// @Tags API.log
	// @Router /api/log/other/add [post]
	AddOtherTable() core.HandlerFunc

	// DeleteOtherTable 移除外部日志表
	// @Tags API.log
	// @Router /api/log/other/delete [delete]
	DeleteOtherTable() core.HandlerFunc

	// QueryLogContext 获取日志上下文
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
	_, err := logservice.InitParseLogTable(req)
	if err != nil {
		logger.Error("create default log table failed", zap.Error(err))
	}
	return &handler{
		logger:      logger,
		logService:  logservice,
		dataService: data.New(dbRepo, promRepo, k8sApi),
	}
}
