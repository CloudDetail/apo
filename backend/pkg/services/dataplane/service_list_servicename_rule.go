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
	for i := range serviceRuleConditions {
		condition := serviceRuleConditions[i]
		conditionMap[condition.RuleID] = append(conditionMap[condition.RuleID], &condition)
	}

	ruleIdStr := make([]string, 0, len(ruleIds))
	for _, ruleId := range ruleIds {
		ruleIdStr = append(ruleIdStr, strconv.Itoa(ruleId))
	}
	apps, err := s.chRepo.GetResolvedApps(ctx, ruleIdStr, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	instanceMapByRuleID := make(map[string]map[string]struct{})
	for _, app := range apps {
		ruleID, ok := app.Labels["rule_id"]
		if !ok {
			continue
		}

		if _, exist := instanceMapByRuleID[ruleID]; !exist {
			instanceMapByRuleID[ruleID] = make(map[string]struct{})
		}

		instanceName := app.GetInstanceName(req.Service)
		instanceMapByRuleID[ruleID][instanceName] = struct{}{}
	}

	rules := make([]*response.ListServiceNameRule, 0, len(serviceRules))
	for _, serviceRule := range serviceRules {
		conditions := conditionMap[serviceRule.ID]

		instanceSet, ok := instanceMapByRuleID[strconv.Itoa(serviceRule.ID)]
		var instances []string
		if ok && len(instanceSet) > 0 {
			instances = make([]string, 0, len(instanceSet))
			for name := range instanceSet {
				instances = append(instances, name)
			}
		} else {
			instances = make([]string, 0)
		}

		rules = append(rules, &response.ListServiceNameRule{
			Id:          serviceRule.ID,
			ServiceName: serviceRule.Service,
			Conditions:  conditions,
			ClusterId:   serviceRule.ClusterId,
			Instances:   instances,
		})
	}

	return &response.ListServiceNameRuleResponse{
		Rules: rules,
	}, nil
}
