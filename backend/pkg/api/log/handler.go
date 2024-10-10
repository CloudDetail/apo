package log

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
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

	// CreateLogTable 创建日志表
	// @Tags API.log
	// @Router /api/log/create [post]
	CreateLogTable() core.HandlerFunc

	DropLogTable() core.HandlerFunc

	UpdateLogTable() core.HandlerFunc

	//GetLogTableInfo() core.HandlerFunc
	QueryLog() core.HandlerFunc

	GetLogChart() core.HandlerFunc
}

type handler struct {
	logger     *zap.Logger
	logService log.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, dbRepo database.Repo) Handler {
	return &handler{
		logger:     logger,
		logService: log.New(chRepo, dbRepo),
	}
}
