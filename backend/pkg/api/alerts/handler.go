package alerts

import (
	"go.uber.org/zap"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/alerts"
)

type Handler interface {
	// InputAlertManager 获取 AlertManager 的告警事件
	// @Tags API.alerts
	// @Router /api/alerts/inputs/alertmanager [post]
	InputAlertManager() core.HandlerFunc
}

type handler struct {
	logger       *zap.Logger
	alertService alerts.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo) Handler {
	return &handler{
		logger:       logger,
		alertService: alerts.New(chRepo),
	}
}
