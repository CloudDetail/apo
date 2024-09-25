package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetAlertRules(req *request.GetAlertRuleRequest) response.GetAlertRulesResponse {
	if req.PageParam == nil {
		req.PageParam = &request.PageParam{
			CurrentPage: 1,
			PageSize:    999,
		}
	}

	rules, totalCount := s.k8sApi.GetAlertRules(req.AlertRuleFile, req.AlertRuleFilter, req.PageParam, req.RefreshCache)
	return response.GetAlertRulesResponse{
		AlertRules: rules,
		Pagination: &model.Pagination{
			Total:       int64(totalCount),
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
		},
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
