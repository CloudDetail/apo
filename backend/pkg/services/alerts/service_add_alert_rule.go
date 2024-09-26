package alerts

import (
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) AddAlertRule(req *request.AddAlertRuleRequest) error {
	if !checkOrFillGroupsLabel(req.AlertRule.Group, req.AlertRule.Labels) {
		return fmt.Errorf("group name and group label mismatch")
	}

	return s.k8sApi.AddAlertRule(req.AlertRuleFile, req.AlertRule)
}
