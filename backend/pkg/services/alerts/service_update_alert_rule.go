package alerts

import "github.com/CloudDetail/apo/backend/pkg/model/request"

// UpdateAlertRule implements Service.
func (s *service) UpdateAlertRule(req *request.UpdateAlertRuleRequest) error {
	return s.k8sApi.UpdateAlertManagerRule(req.AlertRules)
}
