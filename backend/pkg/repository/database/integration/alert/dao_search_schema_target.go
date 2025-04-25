// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (repo *subRepo) SearchSchemaTarget(
	schema string,
	sourceField string, sourceValue string,
	targets []alert.AlertEnrichSchemaTarget,
) ([]string, error) {
	schema = SchemaPrefix + schema

	rows, err := repo.db.Table(schema).Where(sourceField, sourceValue).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()

	var targetsValues = make([]string, 0, len(targets))

	if rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		for _, target := range targets {
			var fieldValue = ""
			for i, column := range columns {
				if target.SchemaField == column {
					fieldValue = values[i].(string)
					break
				}
			}
			targetsValues = append(targetsValues, fieldValue)
		}
	}
	return targetsValues, nil
}
