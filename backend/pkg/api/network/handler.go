package deepflow

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/network"
	"go.uber.org/zap"
)

type Handler interface {
	// GetPodMap 查询 Pod 网络调用拓扑与调用指标
	// @Tags API.Network
	// @Router /api/network/podmap [get]
	GetPodMap() core.HandlerFunc
	// GetSpanSegmentsMetrics 客户端对外调用Span网络耗时分段指标
	// @Tags API.Network
	// @Router /api/network/segments [get]
	GetSpanSegmentsMetrics() core.HandlerFunc
}

type handler struct {
	logger         *zap.Logger
	networkService network.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo) Handler {
	return &handler{
		logger:         logger,
		networkService: network.New(chRepo),
	}
}
