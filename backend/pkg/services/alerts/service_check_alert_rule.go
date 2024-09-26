package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) CheckAlertRule(req *request.CheckAlertRuleRequest) (response.CheckAlertRuleResponse, error) {
	var resp response.CheckAlertRuleResponse
	find, err := s.k8sApi.CheckAlertRule(req.AlertRuleFile, req.Group, req.Alert)
	if err != nil {
		resp.Available = false
		return resp, err
	}

	resp.Available = find
	return resp, nil
}
