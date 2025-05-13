// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) AddAlertRule(ctx_core core.Context, req *request.AddAlertRuleRequest) error {
	if !checkOrFillGroupsLabel(req.AlertRule.Group, req.AlertRule.Labels) {
		return core.Error(code.AlertGroupAndLabelMismatchError, "gourp and group label mismatch")
	}

	return s.k8sApi.AddAlertRule(req.AlertRuleFile, req.AlertRule)
}
