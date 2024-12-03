package alertanalyze

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	// ========================告警分析========================

	// AlertImpact 获取告警事件的影响面
	// 如果关联所需的Label不足,error会返回ErrAlertImpactMissingTag提示期望哪些tag
	AlertImpact(req *request.AlertImpactRequest) ([]clickhouse.EntryRelationship, []response.ImpactAlertEvent, map[model.EndpointKey]int, *model.Pagination, error)

	// 获取告警分析时的上下游调用关系
	GetServiceEndpointRelationForAnalyze(req *request.GetAlertServiceEndpointRelationRequest) (*response.GetServiceEndpointRelationResponse, error)

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
	DetectDefectsOptions() (*response.GetPredefinedDetectExprResponse, error)

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
}

func New(chRepo clickhouse.Repo, promRepo prometheus.Repo, polRepo polarisanalyzer.Repo, k8sRepo kubernetes.Repo, dbRepo database.Repo) Service {
	return &service{
		chRepo:   chRepo,
		promRepo: promRepo,
		polRepo:  polRepo,
		k8sRepo:  k8sRepo,
		dbRepo:   dbRepo,
	}
}
