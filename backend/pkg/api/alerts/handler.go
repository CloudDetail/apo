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

	// GetAlertRuleFile 获取基础告警规则
	// @Tags API.alerts
	// @Router /api/alerts/rules/file [get]
	GetAlertRuleFile() core.HandlerFunc

	// UpdateAlertRuleFile 更新基础告警规则
	// @Tags API.alerts
	// @Router /api/alerts/rules/file [post]
	UpdateAlertRuleFile() core.HandlerFunc

	// GetAlertRules 列出告警规则
	// @Tags API.alerts
	// @Router /api/alerts/rule/list [post]
	GetAlertRules() core.HandlerFunc

	// UpdateAlertRule 更新告警规则
	// @Tags API.alerts
	// @Router /api/alerts/rule [post]
	UpdateAlertRule() core.HandlerFunc

	// DeleteAlertRule 删除告警规则
	// @Tags API.alerts
	// @Router /api/alerts/rule [delete]
	DeleteAlertRule() core.HandlerFunc

	// GetAlertManagerConfigReceiver 列出告警通知对象
	// @Tags API.alerts
	// @Router /api/alerts/alertmanager/receiver/list [post]
	GetAlertManagerConfigReceiver() core.HandlerFunc

	// AddAlertManagerConfigReceiver 新增告警通知对象
	// @Tags API.alerts
	// @Router /api/alerts/alertmanager/receiver/add [post]
	AddAlertManagerConfigReceiver() core.HandlerFunc

	// UpdateAlertManagerConfigReceiver 更新告警通知对象
	// @Tags API.alerts
	// @Router /api/alerts/alertmanager/receiver [post]
	UpdateAlertManagerConfigReceiver() core.HandlerFunc

	// DeleteAlertManagerConfigReceiver 删除告警通知对象
	// @Tags API.alerts
	// @Router /api/alerts/alertmanager/receiver [delete]
	DeleteAlertManagerConfigReceiver() core.HandlerFunc
	// GetGroupList 获取group和label对应接口
	// @Tags API.alerts
	// @Router /api/alerts/rule/groups [get]
	GetGroupList() core.HandlerFunc

	// GetMetricPQL 获取告警规则中指标和PQL
	// @Tags API.alerts
	// @Router /api/alerts/rule/metrics [get]
	GetMetricPQL() core.HandlerFunc

	// AddAlertRule 新增告警规则
	// @Tags API.alerts
	// @Router /api/alerts/rule/add [post]
	AddAlertRule() core.HandlerFunc

	// CheckAlertRule 检查告警规则名是否可用
	// @Tags API.alerts
	// @Router /api/alerts/rule/available/file/group/alert [get]
	CheckAlertRule() core.HandlerFunc
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
