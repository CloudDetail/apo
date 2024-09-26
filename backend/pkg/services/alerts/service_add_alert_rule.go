package alerts

import (
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) AddAlertRule(req *request.AddAlertRuleRequest) error {
	if !checkOrFillGroupsLabel(req.AlertRule.Group, req.AlertRule.Labels) {
		return model.ErrorWithMessage{
			Err:  fmt.Errorf("gourp and group label mismatch"),
			Code: code.AlertGroupAndLabelMismatchError,
		}
	}

	return s.k8sApi.AddAlertRule(req.AlertRuleFile, req.AlertRule)
}
