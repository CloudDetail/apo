package metric

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/metric"
	"go.uber.org/zap"
)

type Handler interface {

	// ListMetrics
	// @Tags API.metric
	// @Router /api/metric/list [get]
	ListMetrics() core.HandlerFunc

	// QueryMetrics
	// @Tags API.metric
	// @Router /api/metric/query [post]
	QueryMetrics() core.HandlerFunc
}

type handler struct {
	logger        *zap.Logger
	metricService metric.Service
}

func New(logger *zap.Logger, promRepo prometheus.Repo) Handler {
	return &handler{
		logger:        logger,
		metricService: metric.New(promRepo),
	}
}
