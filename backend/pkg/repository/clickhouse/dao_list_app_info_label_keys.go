// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/core"
)

const (
	SQL_LIST_APP_INFO_LABEL_KEYS = `SELECT DISTINCT arrayJoin(mapKeys(labels)) AS key
FROM originx_app_info
%s`
	SQL_LIST_APP_INFO_LABEL_VALUES = `SELECT DISTINCT %s
FROM originx_app_info
%s`
)

func (repo *chRepo) ListAppInfoLabelKeys(ctx core.Context, startTime, endTime int64) ([]string, error) {
	qb := NewQueryBuilder().
		NotGreaterThan("start_time", endTime/1000000).
		NotLessThan("heart_time", startTime/1000000)

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

var safeKeyRe = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

func escapeAppInfoKeys(key string) (string, error) {
	if !safeKeyRe.MatchString(key) {
		return "", fmt.Errorf("invalid label key")
	}
	escapedKey := strings.ReplaceAll(key, `'`, `''`)
	return "labels['" + escapedKey + "']", nil
}

func (repo *chRepo) ListAppInfoLabelValues(ctx core.Context, startTime, endTime int64, key string) ([]string, error) {
	labelKey, err := escapeAppInfoKeys(key)
	if err != nil {
		return nil, err
	}
	fb := NewFieldBuilder().
		Alias(labelKey, "value")

	qb := NewQueryBuilder().
		NotGreaterThan("start_time", endTime/1000000).
		NotLessThan("heart_time", startTime/1000000)

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
