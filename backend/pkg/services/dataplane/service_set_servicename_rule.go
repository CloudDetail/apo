package dataplane

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) SetServiceNameRule(ctx core.Context, req *request.SetServiceNameRuleRequest) error {
	if req.ServiceName == "" {
		return errors.New("serviceName is miss")
	}

	ruleId := req.RuleId
	if ruleId == 0 {
		rule := &database.ServiceNameRule{
			Service:   req.ServiceName,
			ClusterId: req.ClusterId,
		}
		if err := s.dbRepo.CreateServiceNameRule(ctx, rule); err != nil {
			return err
		}
		ruleId = rule.ID
	}

	for _, condition := range req.Conditions {
		ruleCondition := &database.ServiceNameRuleCondition{
			ID: condition.CondtiondId,
			RuleID: ruleId,
			Key: condition.Key,
			MatchType: condition.MatchType,
			Value: condition.Value,
		}
		if err := s.dbRepo.UpsertServiceNameRuleCondition(ctx, ruleCondition); err != nil {
			return err
		}
	}
	return nil
}