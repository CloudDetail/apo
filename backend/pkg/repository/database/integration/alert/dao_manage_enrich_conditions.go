// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (repo *subRepo) AddAlertEnrichConditions(ctx_core core.Context, enrichConditions []alert.AlertEnrichCondition) error {
	if len(enrichConditions) == 0 {
		return nil
	}
	return repo.db.Create(&enrichConditions).Error
}

func (repo *subRepo) GetAlertEnrichConditions(ctx_core core.Context, sourceId string) ([]alert.AlertEnrichCondition, error) {
	var enrichConditions []alert.AlertEnrichCondition
	err := repo.db.Find(&enrichConditions, "source_id = ?", sourceId).Error
	return enrichConditions, err
}

func (repo *subRepo) DeleteAlertEnrichConditions(ctx_core core.Context, ruleIds []string) error {
	if len(ruleIds) == 0 {
		return nil
	}
	return repo.db.Delete(&alert.AlertEnrichCondition{}, "enrich_rule_id in ?", ruleIds).Error
}

func (repo *subRepo) DeleteAlertEnrichConditionsBySourceId(ctx_core core.Context, sourceId string) error {
	return repo.db.Delete(&alert.AlertEnrichCondition{}, "source_id = ?", sourceId).Error
}
