package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
)

var _ Service = (*service)(nil)

type Service interface {
	// InputAlertManager 接收 AlertManager 的告警事件
	InputAlertManager(req *request.InputAlertManagerRequest) error

	// GetAlertRule 获取基础告警规则
	GetAlertRule(req *request.GetAlertRuleRequest) (*response.GetAlertRuleResponse, error)

	// UpdateAlertRule 更新告警基础规则
	UpdateAlertRule(req *request.UpdateAlertRuleRequest) error
}

type service struct {
	chRepo clickhouse.Repo
	k8sApi kubernetes.Repo
}



func New(chRepo clickhouse.Repo, k8sApi kubernetes.Repo) Service {
	return &service{
		chRepo: chRepo,
		k8sApi: k8sApi,
	}
}
