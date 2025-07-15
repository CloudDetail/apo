// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) DeleteServiceNameRule(ctx core.Context, req *request.DeleteServiceNameRuleRequest) error {
	exists, err := s.dbRepo.ServiceNameRuleExists(ctx, req.RuleId)
	if err != nil {
		return err
	}

	if !exists {
		return core.Error(code.RoleNotExistsError, "service name rule does not exist")
	}

	return s.dbRepo.DeleteServiceNameRule(ctx, req.RuleId)
}
