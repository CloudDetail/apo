// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) CheckServiceNameRule(ctx core.Context, req *request.SetServiceNameRuleRequest) (*response.CheckServiceNameRuleResponse, error) {
	apps, err := s.chRepo.GetToResolveApps(ctx)
	if err != nil {
		return nil, err
	}

	conditions := make([]*database.ServiceNameRuleCondition, 0)
	for _, condition := range req.Conditions {
		conditions = append(conditions, &database.ServiceNameRuleCondition{
			Key:       condition.Key,
			MatchType: condition.MatchType,
			Value:     condition.Value,
		})
	}

	instances := map[string]bool{}
	result := make([]*model.MatchServiceInstance, 0)
	for _, app := range apps {
		if checkMatch(&app, conditions) {
			instanceName := app.GetInstanceName(req.ServiceName)
			if _, ok := instances[instanceName]; !ok {
				instances[instanceName] = true
				result = append(result, model.NewMatchServiceInstance(req.ServiceName, &app))
			}
		}
	}
	return &response.CheckServiceNameRuleResponse{
		Instances: result,
	}, nil
}

func checkMatch(app *model.AppInfo, conditions []*database.ServiceNameRuleCondition) bool {
	if len(conditions) == 0 {
		return false
	}
	for _, condition := range conditions {
		if !condition.Match(app.Labels) {
			return false
		}
	}
	return true
}
