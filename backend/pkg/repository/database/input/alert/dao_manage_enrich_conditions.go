// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/CloudDetail/apo/backend/pkg/model/input/alert"

func (repo *subRepo) AddAlertEnrichConditions(enrichConditions []alert.AlertEnrichCondition) error {
	if len(enrichConditions) == 0 {
		return nil
	}
	return repo.db.Create(&enrichConditions).Error
}

func (repo *subRepo) GetAlertEnrichConditions(sourceId string) ([]alert.AlertEnrichCondition, error) {
	var enrichConditions []alert.AlertEnrichCondition
	err := repo.db.Find(&enrichConditions, "source_id = ?", sourceId).Error
	return enrichConditions, err
}

func (repo *subRepo) DeleteAlertEnrichConditions(ruleIds []string) error {
	if len(ruleIds) == 0 {
		return nil
	}
	return repo.db.Delete(&alert.AlertEnrichCondition{}, "enrich_rule_id in ?", ruleIds).Error
}

func (repo *subRepo) DeleteAlertEnrichConditionsBySourceId(sourceId string) error {
	return repo.db.Delete(&alert.AlertEnrichCondition{}, "source_id = ?", sourceId).Error
}
