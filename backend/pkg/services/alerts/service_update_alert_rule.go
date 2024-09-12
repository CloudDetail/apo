package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) UpdateAlertRule(req *request.UpdateAlertRuleRequest) error {
	return s.k8sApi.AddOrUpdateAlertRule(req.AlertRuleFile, req.AlertRule)
}

func (s *service) DeleteAlertRule(req *request.DeleteAlertRuleRequest) error {
	return s.k8sApi.DeleteAlertRule(req.AlertRuleFile, req.Group, req.Alert)
}

func (s *service) UpdateAlertRuleFile(req *request.UpdateAlertRuleConfigRequest) error {
	return s.k8sApi.UpdateAlertRuleConfigFile(req.AlertRuleFile, []byte(req.Content))
}
