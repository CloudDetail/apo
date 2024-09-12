package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetAlertRules(configFile string) response.GetAlertRulesResponse {
	return response.GetAlertRulesResponse{
		AlertRules: s.k8sApi.GetAlertRules(configFile),
	}
}

// GetAlertRuleFile 获取告警规则
func (s *service) GetAlertRuleFile(req *request.GetAlertRuleConfigRequest) (*response.GetAlertRuleFileResponse, error) {
	rules, err := s.k8sApi.GetAlertRuleConfigFile(req.AlertRuleFile)
	if err != nil {
		return &response.GetAlertRuleFileResponse{AlertRules: map[string]string{}}, err
	}

	return &response.GetAlertRuleFileResponse{AlertRules: rules}, nil
}
