// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/core"
)

const SQL_LIST_APP_INFO_LABEL_KEYS = `SELECT DISTINCT arrayJoin(mapKeys(labels)) AS key
FROM originx_app_info
%s`

const SQL_LIST_APP_INFO_LABEL_VALUES = `SELECT DISTINCT %s AS value
FROM originx_app_info
%s`

func (repo *chRepo) ListAppInfoLabelKeys(ctx core.Context, startTime, endTime int64) ([]string, error) {
	qb := NewQueryBuilder().
		Between("timestamp", startTime/1e6, endTime/1e6)

	sql := fmt.Sprintf(SQL_LIST_APP_INFO_LABEL_KEYS, qb.String())
	rows, err := repo.GetConn(ctx).Query(ctx.GetContext(), sql, qb.values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var result []string
	for rows.Next() {
		var labelsKey string
		if err := rows.Scan(&labelsKey); err != nil {
			return nil, err
		}
		result = append(result, labelsKey)
	}
	return result, nil
}

func escapeAppInfoKeys(key string) string {
	escapedKey := strings.ReplaceAll(key, `'`, `''`)
	expr := `labels['` + escapedKey + `']`
	escapedExpr := strings.ReplaceAll(expr, "`", "``")
	return "`" + escapedExpr + "`"
}

func (repo *chRepo) ListAppInfoLabelValues(ctx core.Context, startTime, endTime int64, key string) ([]string, error) {
	fb := NewFieldBuilder().
		Alias(escapeAppInfoKeys(key), "value")

	qb := NewQueryBuilder().
		Between("timestamp", startTime/1e6, endTime/1e6)

	sql := fmt.Sprintf(SQL_LIST_APP_INFO_LABEL_VALUES, fb.String(), qb.String())
	rows, err := repo.GetConn(ctx).Query(ctx.GetContext(), sql, qb.values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var result []string
	for rows.Next() {
		var labelsValue string
		if err := rows.Scan(&labelsValue); err != nil {
			return nil, err
		}
		result = append(result, labelsValue)
	}
	return result, nil
}
