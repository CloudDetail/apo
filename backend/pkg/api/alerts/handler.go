package alerts

import (
	"go.uber.org/zap"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/services/alerts"
)

type Handler interface {
	// InputAlertManager 获取 AlertManager 的告警事件
	// @Tags API.alerts
	// @Router /api/alerts/inputs/alertmanager [post]
	InputAlertManager() core.HandlerFunc

	// GetAlertRule 获取基础告警规则
	// @Tags API.alerts
	// @Router /api/alerts/rules [get]
	GetAlertRule() core.HandlerFunc

	// UpdateAlertRule 更新基础告警规则
	// @Tags API.alerts
	// @Router /api/alerts/rules [post]
	UpdateAlertRule() core.HandlerFunc
}

type handler struct {
	logger       *zap.Logger
	alertService alerts.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, k8sRepo kubernetes.Repo) Handler {
	return &handler{
		logger:       logger,
		alertService: alerts.New(chRepo, k8sRepo),
	}
}
