// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"strconv"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) ListServiceNameRule(ctx core.Context, req *request.ListServiceNameRuleRequest) (*response.ListServiceNameRuleResponse, error) {
	serviceRules, err := s.dbRepo.ListServiceNameRule(ctx, req.Service)
	if err != nil {
		return nil, err
	}
	if len(serviceRules) == 0 {
		return &response.ListServiceNameRuleResponse{
			Rules: []*response.ListServiceNameRule{},
		}, nil
	}

	ruleIds := make([]int, len(serviceRules))
	for i := range serviceRules {
		ruleIds[i] = serviceRules[i].ID
	}
	serviceRuleConditions, err := s.dbRepo.ListServiceNameRuleCondition(ctx, ruleIds)
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
		conditions, ok := conditionMap[serviceRule.ID]
		if !ok {
			conditions = []*database.ServiceNameRuleCondition{}
		}
		rules = append(rules, &response.ListServiceNameRule{
			Id:          serviceRule.ID,
			ServiceName: serviceRule.Service,
			Conditions:  conditions,
			ClusterId:   serviceRule.ClusterId,
		})
	}

	ruleIdStr := make([]string, 0, len(ruleIds))
	for _, ruleId := range ruleIds {
		ruleIdStr = append(ruleIdStr, strconv.Itoa(ruleId))
	}
	apps, err := s.chRepo.GetResolvedApps(ctx, ruleIdStr)
	if err != nil {
		return nil, err
	}

	instanceMap := make(map[string]struct{})
	instanceNames := make([]string, 0, len(apps))
	for _, app := range apps {
		instanceName := app.GetInstanceName(req.Service)
		if _, exist := instanceMap[instanceName]; !exist {
			instanceNames = append(instanceNames, instanceName)
			instanceMap[instanceName] = struct{}{}
		}
	}

	return &response.ListServiceNameRuleResponse{
		Rules:     rules,
		Instances: instanceNames,
	}, nil
}
