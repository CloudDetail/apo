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
	result := make([]*model.AppInfo, 0)
	for _, app := range apps {
		if req.ClusterId == app.Labels["cluster_id"] && checkMatch(app, conditions) {
			result = append(result, app)
		}
	}
	return &response.CheckServiceNameRuleResponse{
		Apps: result,
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
