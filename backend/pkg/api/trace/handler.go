package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/trace"
	"go.uber.org/zap"
)

type Handler interface {
	// GetTracePageList 查询Trace分页列表
	// @Tags API.trace
	// @Router /api/trace/pagelist [post]
	GetTracePageList() core.HandlerFunc
}

type handler struct {
	logger       *zap.Logger
	traceService trace.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo) Handler {
	return &handler{
		logger:       logger,
		traceService: trace.New(chRepo),
	}
}
