package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetAlertRule 获取告警规则
func (s *service) GetAlertRule(req *request.GetAlertRuleRequest) (*response.GetAlertRuleResponse, error) {
	rules, err := s.k8sApi.GetAlertManagerRule(req.AlertRuleFile)
	if err != nil {
		return &response.GetAlertRuleResponse{AlertRules: map[string]string{}}, err
	}

	return &response.GetAlertRuleResponse{AlertRules: rules}, nil
}
