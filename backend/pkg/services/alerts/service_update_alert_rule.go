// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
)

func (s *service) UpdateAlertRule(req *request.UpdateAlertRuleRequest) error {
	if !checkOrFillGroupsLabel(req.AlertRule.Group, req.AlertRule.Labels) {
		return model.NewErrWithMessage(
			fmt.Errorf("gourp and group label mismatch"),
			code.AlertGroupAndLabelMismatchError)
	}

	return s.k8sApi.UpdateAlertRule(req.AlertRuleFile, req.AlertRule, req.OldGroup, req.OldAlert)
}

func (s *service) DeleteAlertRule(req *request.DeleteAlertRuleRequest) error {
	return s.k8sApi.DeleteAlertRule(req.AlertRuleFile, req.Group, req.Alert)
}

func (s *service) UpdateAlertRuleFile(req *request.UpdateAlertRuleConfigRequest) error {
	return s.k8sApi.UpdateAlertRuleConfigFile(req.AlertRuleFile, []byte(req.Content))
}

// checkOrFillGroupsLabel check the correspondence between the group and the label. if the label is empty, fill it
func checkOrFillGroupsLabel(group string, labels map[string]string) bool {
	groupLabel := labels["group"]
	label, ok := kubernetes.GetLabel(group)
	if ok && groupLabel != label {
		return false
	}

	labels["group"] = label

	return true
}
