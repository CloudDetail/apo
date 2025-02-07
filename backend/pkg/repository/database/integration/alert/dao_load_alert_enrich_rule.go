// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"

func (repo *subRepo) LoadAlertEnrichRule() ([]alert.AlertSource, map[alert.SourceFrom][]alert.AlertEnrichRuleVO, error) {
	var sources []alert.AlertSource
	err := repo.db.Find(&sources).Error
	if err != nil {
		return nil, nil, err
	}

	var enrichRules []alert.AlertEnrichRule
	err = repo.db.Find(&enrichRules).Error
	if err != nil {
		return nil, nil, err
	}

	var enrichConditions []alert.AlertEnrichCondition
	err = repo.db.Find(&enrichConditions).Error
	if err != nil {
		return nil, nil, err
	}

	var enrichSchemaTarget []alert.AlertEnrichSchemaTarget
	err = repo.db.Find(&enrichSchemaTarget).Error
	if err != nil {
		return nil, nil, err
	}

	var rules = make(map[alert.SourceFrom][]alert.AlertEnrichRuleVO)
	for _, source := range sources {
		var ruleVO []alert.AlertEnrichRuleVO
		for _, rule := range enrichRules {
			if rule.SourceID == source.SourceID {
				ruleVO = append(ruleVO, alert.AlertEnrichRuleVO{
					AlertEnrichRule: rule,
					Conditions:      searchConditions(rule.EnrichRuleID, enrichConditions),
					SchemaTargets:   searchSchemaTarget(rule.EnrichRuleID, enrichSchemaTarget),
				})
			}
		}
		rules[source.SourceFrom] = ruleVO
	}

	return sources, rules, nil
}

func searchConditions(ruleId string, conditions []alert.AlertEnrichCondition) []alert.AlertEnrichCondition {
	var res = make([]alert.AlertEnrichCondition, 0)
	for _, condition := range conditions {
		if condition.EnrichRuleID == ruleId {
			res = append(res, condition)
		}
	}
	return res
}

func searchSchemaTarget(ruleId string, enrichSchemaTarget []alert.AlertEnrichSchemaTarget) []alert.AlertEnrichSchemaTarget {
	var res = make([]alert.AlertEnrichSchemaTarget, 0)
	for _, schemaTarget := range enrichSchemaTarget {
		if schemaTarget.EnrichRuleID == ruleId {
			res = append(res, schemaTarget)
		}
	}
	return res
}
