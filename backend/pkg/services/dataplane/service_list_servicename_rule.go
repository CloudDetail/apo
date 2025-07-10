// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) ListServiceNameRule(ctx core.Context) (*response.ListServiceNameRuleResponse, error) {
	serviceRules, err := s.dbRepo.ListAllServiceNameRule(ctx)
	if err != nil {
		return nil, err
	}
	if len(serviceRules) == 0 {
		return &response.ListServiceNameRuleResponse{}, nil
	}

	serviceRuleConditions, err := s.dbRepo.ListAllServiceNameRuleCondition(ctx)
	if err != nil {
		return nil, err
	}
	conditionMap := make(map[int][]*database.ServiceNameRuleCondition)
	for _, condition := range serviceRuleConditions {
		existConditions, exist := conditionMap[condition.RuleID]
		if !exist {
			existConditions = make([]*database.ServiceNameRuleCondition, 0)
		}
		existConditions = append(existConditions, &condition)
		conditionMap[condition.RuleID] = existConditions
	}

	rules := make([]*response.ListServiceNameRule, 0)
	for _, serviceRule := range serviceRules {
		conditions := conditionMap[serviceRule.ID]
		rules = append(rules, &response.ListServiceNameRule{
			Id:          serviceRule.ID,
			ServiceName: serviceRule.Service,
			Conditions:  conditions,
		})
	}
	return &response.ListServiceNameRuleResponse{
		Rules: rules,
	}, nil
}
