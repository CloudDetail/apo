// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (repo *subRepo) AddAlertEnrichRule(ctx core.Context, enrichRule []alert.AlertEnrichRule) error {
	if len(enrichRule) == 0 {
		return nil
	}
	return repo.db.Create(&enrichRule).Error
}

func (repo *subRepo) GetAlertEnrichRule(ctx core.Context, sourceId string) ([]alert.AlertEnrichRule, error) {
	var enrichRules []alert.AlertEnrichRule
	err := repo.db.Find(&enrichRules, "source_id = ?", sourceId).Error
	return enrichRules, err
}

func (repo *subRepo) DeleteAlertEnrichRule(ctx core.Context, ruleIds []string) error {
	if len(ruleIds) == 0 {
		return nil
	}
	return repo.db.Delete(&alert.AlertEnrichRule{}, "enrich_rule_id in ?", ruleIds).Error
}

func (repo *subRepo) DeleteAlertEnrichRuleBySourceId(ctx core.Context, sourceId string) error {
	return repo.db.Delete(&alert.AlertEnrichRule{}, "source_id = ?", sourceId).Error
}
