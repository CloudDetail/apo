// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (repo *subRepo) AddAlertEnrichConditions(ctx core.Context, enrichConditions []alert.AlertEnrichCondition) error {
	if len(enrichConditions) == 0 {
		return nil
	}
	return repo.GetContextDB(ctx).Create(&enrichConditions).Error
}

func (repo *subRepo) GetAlertEnrichConditions(ctx core.Context, sourceId string) ([]alert.AlertEnrichCondition, error) {
	var enrichConditions []alert.AlertEnrichCondition
	err := repo.GetContextDB(ctx).Find(&enrichConditions, "source_id = ?", sourceId).Error
	return enrichConditions, err
}

func (repo *subRepo) DeleteAlertEnrichConditions(ctx core.Context, ruleIds []string) error {
	if len(ruleIds) == 0 {
		return nil
	}
	return repo.GetContextDB(ctx).Delete(&alert.AlertEnrichCondition{}, "enrich_rule_id in ?", ruleIds).Error
}

func (repo *subRepo) DeleteAlertEnrichConditionsBySourceId(ctx core.Context, sourceId string) error {
	return repo.GetContextDB(ctx).Delete(&alert.AlertEnrichCondition{}, "source_id = ?", sourceId).Error
}
