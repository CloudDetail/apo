// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"

func (repo *subRepo) CheckSchemaIsUsed(schema string) ([]string, error) {
	var alertSource = make([]string, 0)
	if !AllowSchema.MatchString(schema) {
		return nil, alert.ErrNotAllowSchema{Table: schema}
	}

	sql := `SELECT source_name,schema FROM alert_enrich_rules st left join alert_sources s on s.source_id = st.source_id WHERE st.schema = ?`
	rows, err := repo.db.Raw(sql, schema).Rows()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	columns, _ := rows.Columns()
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		alertSource = append(alertSource, values[0].(string))
	}
	return alertSource, err
}

func (repo *subRepo) AddAlertEnrichSchemaTarget(enrichSchemaTarget []alert.AlertEnrichSchemaTarget) error {
	if len(enrichSchemaTarget) == 0 {
		return nil
	}
	return repo.db.Create(&enrichSchemaTarget).Error
}

func (repo *subRepo) GetAlertEnrichSchemaTarget(sourceId string) ([]alert.AlertEnrichSchemaTarget, error) {
	var enrichSchemaTarget []alert.AlertEnrichSchemaTarget
	err := repo.db.Find(&enrichSchemaTarget, "source_id = ?", sourceId).Error
	return enrichSchemaTarget, err
}

func (repo *subRepo) DeleteAlertEnrichSchemaTarget(ruleIds []string) error {
	if len(ruleIds) == 0 {
		return nil
	}
	return repo.db.Delete(&alert.AlertEnrichSchemaTarget{}, "enrich_rule_id in ?", ruleIds).Error
}

func (repo *subRepo) DeleteAlertEnrichSchemaTargetBySourceId(sourceId string) error {
	return repo.db.Delete(&alert.AlertEnrichSchemaTarget{}, "source_id = ?", sourceId).Error
}
