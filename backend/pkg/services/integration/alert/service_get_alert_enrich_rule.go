// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (s *service) GetAlertEnrichRule(ctx core.Context,
	sourceID string,
) ([]alert.AlertEnrichRuleVO, error) {
	enrichRules, err := s.dbRepo.GetAlertEnrichRule(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	conditions, err := s.dbRepo.GetAlertEnrichConditions(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	schemaTargets, err := s.dbRepo.GetAlertEnrichSchemaTarget(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	var alertEnrichRuleVOs = make([]alert.AlertEnrichRuleVO, 0, len(enrichRules))
	for _, enrichRule := range enrichRules {
		enrichRuleVO := alert.AlertEnrichRuleVO{
			Conditions:    []alert.AlertEnrichCondition{},
			SchemaTargets: []alert.AlertEnrichSchemaTarget{},
		}
		enrichRuleVO.AlertEnrichRule = enrichRule

		for _, condition := range conditions {
			if condition.EnrichRuleID == enrichRule.EnrichRuleID {
				enrichRuleVO.Conditions = append(enrichRuleVO.Conditions, condition)
			}
		}

		for _, schemaTarget := range schemaTargets {
			if schemaTarget.EnrichRuleID == enrichRule.EnrichRuleID {
				enrichRuleVO.SchemaTargets = append(enrichRuleVO.SchemaTargets, schemaTarget)
			}
		}

		alertEnrichRuleVOs = append(alertEnrichRuleVOs, enrichRuleVO)
	}
	return alertEnrichRuleVOs, nil
}
