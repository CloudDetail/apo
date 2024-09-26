package alerts

import (
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) UpdateAlertRule(req *request.UpdateAlertRuleRequest) (*response.UpdateAlertRuleResponse, error) {
	if !CheckOrFillGroupsLabel(req.AlertRule.Group, req.AlertRule.Labels) {
		return nil, fmt.Errorf("group name and group label mismatch")
	}

	if err := s.k8sApi.UpdateAlertRule(req.AlertRuleFile, req.AlertRule); err != nil {
		return nil, err
	}

	return &response.UpdateAlertRuleResponse{Group: req.AlertRule.Group, Alert: req.AlertRule.Alert}, nil
}

func (s *service) DeleteAlertRule(req *request.DeleteAlertRuleRequest) error {
	return s.k8sApi.DeleteAlertRule(req.AlertRuleFile, req.Group, req.Alert)
}

func (s *service) UpdateAlertRuleFile(req *request.UpdateAlertRuleConfigRequest) error {
	return s.k8sApi.UpdateAlertRuleConfigFile(req.AlertRuleFile, []byte(req.Content))
}

// CheckOrFillGroupsLabel 检查group与label的对应关系，如果label为空则填充
func CheckOrFillGroupsLabel(group string, labels map[string]string) bool {
	groupLabel := labels["group"]
	switch group {
	case appLabelVal:
		if groupLabel != "" && groupLabel != appLabelKey {
			return false
		}
		groupLabel = appLabelKey
	case containerLabelVal:
		if groupLabel != "" && groupLabel != containerLabelKey {
			return false
		}
		groupLabel = containerLabelKey
	case netLabelVal:
		if groupLabel != "" && groupLabel != netLabelKey {
			return false
		}
		groupLabel = netLabelKey
	case infraLabelVal:
		if groupLabel != "" && groupLabel != infraLabelKey {
			return false
		}
		groupLabel = infraLabelKey
	default:
		if groupLabel != "" && groupLabel != customLabelKey {
			return false
		}
		groupLabel = customLabelKey
	}

	labels["group"] = groupLabel

	return true
}
