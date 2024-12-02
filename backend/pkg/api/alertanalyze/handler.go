package alertanalyze

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	polaris "github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/alertanalyze"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
	"go.uber.org/zap"
)

type Handler interface {
	// ========================告警分析========================

	// GetAlertImpact 获取告警数据的影响面
	// @Tags API.alerts
	// @Router /api/alerts/event/impact [get]
	GetAlertImpact() core.HandlerFunc

	// GetDescendantContribution 获取下游故障贡献度
	// @Tags API.alerts
	// @Router /api/alerts/descendant/anormal/contribution [get]
	GetDescendantContribution() core.HandlerFunc

	// GetDescendantAnormal 获取下游异常
	// @Tags API.alerts
	// @Router /api/alerts/descendant/anormal/list [get]
	GetDescendantAnormal() core.HandlerFunc

	// GetAnormalTrendByEntry 获取下游异常变化趋势
	// @Tags API.alerts
	// @Router /api/alerts/descendant/anormal/trend [get]
	GetAnormalTrendByEntry() core.HandlerFunc

	// GetDescendantAnormalDelta 获取下游异常变化
	// @Tags API.alerts
	// @Router /api/alerts/descendant/anormal/delta [get]
	GetDescendantAnormalDelta() core.HandlerFunc

	// GetAnomalySpan 获取服务和根因类型的故障报告
	// @Tags API.alerts
	// @Router /api/alerts/anomaly-span/list [post]
	GetAnomalySpan() core.HandlerFunc

	// GetPredefinedDetectExpr 获取预定义检测表达式
	// @Tags API.alerts
	// @Router /api/alerts/detect/mutation/metrics [get]
	GetPredefinedDetectExpr() core.HandlerFunc

	// AddDefectsDetect 执行异常检测
	// @Tags API.alerts
	// @Router /api/alerts/detect/mutation/add [post]
	AddDefectsDetect() core.HandlerFunc

	// GetDefectDetectExecList 获取异常检测执行记录
	// @Tags API.alerts
	// @Router /api/alerts/detect/mutation/exec-list [get]
	GetDefectDetectExecList() core.HandlerFunc

	// GetDefectDetectRuleList 获取异常检测规则
	// @Tags API.alerts
	// @Router /api/alerts/detect/mutation/rule-list [get]
	GetDefectDetectRuleList() core.HandlerFunc

	// GetAlertServiceEndpointRelation 获取告警分析中的服务相关性
	// @Tags API.alerts
	// @Router /api/alerts/service/relation [get]
	GetAlertServiceEndpointRelation() core.HandlerFunc
}

type handler struct {
	logger                 *zap.Logger
	alertanalyzeService    alertanalyze.Service
	serviceoverviewService serviceoverview.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo, dbRepo database.Repo, promRepo prometheus.Repo, polRepo polaris.Repo, k8sRepo kubernetes.Repo) Handler {
	return &handler{
		logger:                 logger,
		alertanalyzeService:    alertanalyze.New(chRepo, promRepo, polRepo, k8sRepo, dbRepo),
		serviceoverviewService: serviceoverview.New(chRepo, dbRepo, promRepo),
	}
}
