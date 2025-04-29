// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

const (
	logsBaseQuery = "SELECT * FROM `%s`.`%s` WHERE %s %s"
)

// validateLogQueryRequest validates the LogQueryRequest parameters
func validateLogQueryRequest(req *request.LogQueryRequest) error {
	if req.StartTime <= 0 || req.EndTime <= 0 || req.StartTime > req.EndTime {
		return fmt.Errorf("invalid time range")
	}
	if req.PageSize <= 0 {
		return fmt.Errorf("invalid page size")
	}
	if req.PageNum < 0 {
		return fmt.Errorf("invalid page number")
	}
	return nil
}

func (ch *chRepo) QueryAllLogs(req *request.LogQueryRequest) ([]map[string]any, string, error) {
	// Validate request parameters
	if err := validateLogQueryRequest(req); err != nil {
		return nil, "", fmt.Errorf("invalid request parameters: %w", err)
	}

	if !util.IsValidIdentifier(req.TableName) {
		return nil, "", fmt.Errorf("invalid request parameters: %s", req.TableName)
	}

	if !util.IsValidIdentifier(req.DataBase) {
		return nil, "", fmt.Errorf("invalid request parameters: %s", req.DataBase)
	}

	if !validateQuery(req.Query) {
		return nil, "", fmt.Errorf("invalid request parameters: %s", req.Query)
	}

	// Build safe query condition
	condition := NewQueryCondition(req.StartTime, req.EndTime, req.TimeField, req.Query)

	// Escape SQL identifiers
	timeField := util.EscapeSQLString(req.TimeField)
	database := util.EscapeSQLString(req.DataBase)
	tableName := util.EscapeSQLString(req.TableName)

	// Build order by clause
	bySql := NewByLimitBuilder().
		OrderBy(fmt.Sprintf("`%s`", timeField), false).
		Limit(req.PageSize).
		Offset(req.PageNum).
		String()

	// Build and execute query
	sql := buildAllLogsQuery(logsBaseQuery, database, tableName, condition, bySql)
	results, err := ch.queryRowsData(sql)
	if err != nil {
		return nil, sql, fmt.Errorf("failed to execute query: %w", err)
	}

	return results, sql, nil
}

func buildAllLogsQuery(baseQuery string, database, tableName, condition, bySql string) string {
	return fmt.Sprintf(baseQuery, database, tableName, condition, bySql)
}
