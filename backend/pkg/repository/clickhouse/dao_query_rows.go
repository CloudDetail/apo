// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import core "github.com/CloudDetail/apo/backend/pkg/core"

func (ch *chRepo) queryRowsData(ctx_core core.Context, sql string) ([]map[string]any, error) {
	rows, err := ch.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		entry := make(map[string]interface{})

		for i, col := range columns {
			entry[col] = values[i]
		}
		results = append(results, entry)
	}
	return results, nil
}
