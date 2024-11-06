package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/trace"
	"go.uber.org/zap"
)

type Handler interface {
	// GetTraceFilters 查询Trace列表可用的过滤器
	// @Tags API.trace
	// @Router /api/trace/pagelist/filters [get]
	GetTraceFilters() core.HandlerFunc
	// GetTraceFilterValue 查询指定过滤器的可用数值
	// @Tags API.trace
	// @Router /api/trace/pagelist/filter/value [post]
	GetTraceFilterValue() core.HandlerFunc
	// GetTracePageList 查询Trace分页列表
	// @Tags API.trace
	// @Router /api/trace/pagelist [post]
	GetTracePageList() core.HandlerFunc

	// GetOnOffCPU 获取span执行消耗
	// @Tags API.trace
	// @Router /api/trace/onoffcpu [get]
	GetOnOffCPU() core.HandlerFunc
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
