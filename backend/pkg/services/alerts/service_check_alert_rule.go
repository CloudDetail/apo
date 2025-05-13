// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) CheckAlertRule(ctx_core core.Context, req *request.CheckAlertRuleRequest) (response.CheckAlertRuleResponse, error) {
	var resp response.CheckAlertRuleResponse
	find, err := s.k8sApi.CheckAlertRule(req.AlertRuleFile, req.Group, req.Alert)
	if err != nil {
		resp.Available = false
		return resp, err
	}

	resp.Available = find
	return resp, nil
}
