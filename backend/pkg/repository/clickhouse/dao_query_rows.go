// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (ch *chRepo) queryRowsData(ctx core.Context, sql string) ([]map[string]any, error) {
	rows, err := ch.GetContextDB(ctx).Query(ctx.GetContext(), sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return rowsToMapSlice(rows)
}

func rowsToMapSlice(rows driver.Rows) ([]map[string]any, error) {
	columnNames := rows.Columns()
	columnTypes := rows.ColumnTypes()
	var result []map[string]any

	valuePtrs := make([]any, len(columnNames))
	for i, colType := range columnTypes {
		switch colType.DatabaseTypeName() {
		case "DateTime64", "DateTime", "DateTime64(9)":
			valuePtrs[i] = new(time.Time)
		case "UInt64", "Nullable(UInt64)":
			valuePtrs[i] = new(uint64)
		case "Int64", "Nullable(Int64)":
			valuePtrs[i] = new(int64)
		case "UInt32", "Nullable(Uint32)":
			valuePtrs[i] = new(uint32)
		case "Int32", "Nullable(Int32)":
			valuePtrs[i] = new(int32)
		case "Float64", "Nullable(Float64)":
			valuePtrs[i] = new(float64)
		case "Float32", "Nullable(Float32)":
			valuePtrs[i] = new(float32)
		case "Nullable(String)", "String", "FixedString", "LowCardinality(String)":
			valuePtrs[i] = new(string)
		case "UUID":
			valuePtrs[i] = new(string)
		default:
			// TODO support clickhouse Map
			valuePtrs[i] = new(string)
		}
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]any, len(columnNames))
		for i, name := range columnNames {
			switch val := valuePtrs[i].(type) {
			case *time.Time:
				rowMap[name] = *val
			case *uint64:
				rowMap[name] = *val
			case *int64:
				rowMap[name] = *val
			case *uint32:
				rowMap[name] = *val
			case *int32:
				rowMap[name] = *val
			case *float64:
				rowMap[name] = *val
			case *float32:
				rowMap[name] = *val
			case *string:
				rowMap[name] = *val
			default:
				rowMap[name] = ""
			}
		}
		result = append(result, rowMap)
	}
	return result, nil
}
