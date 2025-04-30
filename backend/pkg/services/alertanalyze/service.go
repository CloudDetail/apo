package alertanalyze

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"go.uber.org/zap"
)

var _ Service = (*service)(nil)

type Service interface {
	// ========================AlertAnalyze========================

	// Get AlertEvents with relatedService and serviceEntry
	// multi events exist for the same alertID
	GetAlertEventsSvc(req *request.GetAlertEventsSvcReq) (*response.GetAlertEventsSvcResp, error)

	GetAlertsImpact(req *request.AlertImpactRequest) ([]response.AlertRelatedEntry, error)

	// 获取下游的故障贡献度排序
	GetDescendantContribution(req *request.GetDescendantAlertContributaionRequest) (resp *response.GetDescendantAlertContributationResponse, err error)

	// GetAnomalySpan 获取可分析的异常Span
	GetAnomalySpan(req *request.GetAnomalySpanRequest) (response.GetAnomalySpanResponse, error)

	// 获取节点的所有下游节点的异常
	GetDescendantAnormal(req *request.GetDescendantAnormalEventRequest) (*response.GetDescendantAnormalEventResponse, error)

	// 获取异常趋势
	GetAnormalTrendByEntry(req *request.GetDescendantAnormalTrendRequest) (*response.GetDescendantDeltaAnormalEventResponse, error)

	// SearchAnormalDeltaByEntry 获取节点的所有下游节点的异常变化
	SearchAnormalDeltaByEntry(req *request.GetDescendantAnormalDeltaEventRequest) (*response.GetDescendantDeltaAnormalEventResponse, error)

	// 获取预定义的检测表达式
	DetectDefectsOptions(ctx core.Context) (*response.GetPredefinedDetectExprResponse, error)

	// 缺陷检测
	DetectDefects(req *request.DetectMutationRequest) error

	GetDefectDetectExecList(req *request.GetDefectDetectExecListRequest) (*response.GetDefectDetectExecListResponse, error)

	GetDefectDetectRuleList(req *request.GetDefectDetectRuleListRequest) (*response.GetDefectDetectRuleListResponse, error)
}

type service struct {
	chRepo   clickhouse.Repo
	promRepo prometheus.Repo
	polRepo  polarisanalyzer.Repo

	k8sRepo kubernetes.Repo
	dbRepo  database.Repo // TODO init

	logger *zap.Logger
}

func New(
	chRepo clickhouse.Repo,
	promRepo prometheus.Repo,
	polRepo polarisanalyzer.Repo,
	k8sRepo kubernetes.Repo,
	dbRepo database.Repo,
	logger *zap.Logger,
) Service {
	return &service{
		chRepo:   chRepo,
		promRepo: promRepo,
		polRepo:  polRepo,
		k8sRepo:  k8sRepo,
		dbRepo:   dbRepo,
		logger:   logger,
	}
}
