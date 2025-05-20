// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
)

func (s *service) UpdateAlertRule(ctx core.Context, req *request.UpdateAlertRuleRequest) error {
	if !checkOrFillGroupsLabel(req.AlertRule.Group, req.AlertRule.Labels) {
		return core.Error(
			code.AlertGroupAndLabelMismatchError,
			"gourp and group label mismatch",
		)
	}

	return s.k8sApi.UpdateAlertRule(req.AlertRuleFile, req.AlertRule, req.OldGroup, req.OldAlert)
}

func (s *service) DeleteAlertRule(ctx core.Context, req *request.DeleteAlertRuleRequest) error {
	return s.k8sApi.DeleteAlertRule(req.AlertRuleFile, req.Group, req.Alert)
}

func (s *service) UpdateAlertRuleFile(ctx core.Context, req *request.UpdateAlertRuleConfigRequest) error {
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
