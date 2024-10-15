package log

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
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

	CreateLogTable() core.HandlerFunc

	DropLogTable() core.HandlerFunc

	// UpdateLogTable 更新日志表
	// @Tags API.log
	// @Router /api/log/update [post]
	UpdateLogTable() core.HandlerFunc

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
	// @Router /api/log/table [post]
	GetLogTableInfo() core.HandlerFunc

	// GetLogParseRule 获取日志表解析规则
	// @Tags API.log
	// @Router /api/log/rule/get [post]
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
	// @Router /api/log/rule/delete [post]
	DeleteLogParseRule() core.HandlerFunc
}

type handler struct {
	logger     *zap.Logger
	logService log.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, dbRepo database.Repo, k8sApi kubernetes.Repo) Handler {
	logservice := log.New(chRepo, dbRepo, k8sApi)
	req := &request.LogTableRequest{}
	req.FillerValue()
	_, err := logservice.CreateLogTable(req)
	if err != nil {
		logger.Error("create default log table failed", zap.Error(err))
	}
	return &handler{
		logger:     logger,
		logService: logservice,
	}
}
