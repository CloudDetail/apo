// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"fmt"
	"regexp"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

var AllowField = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,63}$`)

func (repo *subRepo) SearchSchemaTarget(
	schema string,
	sourceField string, sourceValue string,
	targets []alert.AlertEnrichSchemaTarget,
) ([]string, error) {
	if !AllowSchema.MatchString(schema) {
		return nil, alert.ErrNotAllowSchema{Table: schema}
	}
	schema = SchemaPrefix + schema

	if !AllowField.MatchString(sourceField) {
		return nil, fmt.Errorf("invalid source field: %s", sourceField)
	}
	
	rows, err := repo.db.Table(schema).Where(fmt.Sprintf("%s = ?", sourceField), sourceValue).Limit(1).Rows()
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
