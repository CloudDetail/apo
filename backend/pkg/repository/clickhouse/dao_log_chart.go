// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

// validateLogChartRequest validates the LogQueryRequest parameters for chart queries
func validateLogChartRequest(req *request.LogQueryRequest) error {
	if req.StartTime <= 0 || req.EndTime <= 0 || req.StartTime > req.EndTime {
		return fmt.Errorf("invalid time range")
	}
	if req.TimeField == "" {
		return fmt.Errorf("time field cannot be empty")
	}
	return nil
}

// validateIdentifier checks if a string is a valid SQL identifier
func validateIdentifier(s string) bool {
	// Only allow letters, numbers, underscores, and dots
	validIdentifier := regexp.MustCompile(`^[a-zA-Z0-9_\.]+$`)
	return validIdentifier.MatchString(s)
}

func calculateInterval(interval int64, timeField string) (string, int64, error) {
	// Validate time field
	if !validateIdentifier(timeField) {
		return "", 0, fmt.Errorf("invalid time field: %s", timeField)
	}

	// Escape time field
	escapedTimeField := util.EscapeSQLString(timeField)

	if interval == 0 {
		return "", 0, nil
	}
	if interval <= 60*5 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 1 second)", escapedTimeField), 1, nil
	} else if interval <= 60*30 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 1 minute)", escapedTimeField), 60, nil
	} else if interval <= 60*60*4 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 10 minute)", escapedTimeField), 600, nil
	} else if interval <= 60*60*24 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 1 hour)", escapedTimeField), 3600, nil
	} else if interval <= 60*60*24*7 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 6 hour)", escapedTimeField), 21600, nil
	}
	return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 1 day)", escapedTimeField), 86400, nil
}

const queryLogChart = "SELECT count(`%s`) as count, %s as timeline FROM `%s`.`%s` WHERE %s GROUP BY %s ORDER BY %s ASC"


// HACK input check
func validateInputStr(input string) (string, bool) {
	if len(input) > 100 {
		return "", false
	}
	if !utf8.Valid([]byte(input)) {
		return "", false
	}

	// 使用HTMLEscape处理特殊字符
	escapedResult := template.HTMLEscapeString(input)
	return escapedResult, true
}

// HACK query validation
var dangerousPattern = regexp.MustCompile(`and|exec|execute|insert|select|delete|update|count|drop|chr|mid|master|truncate|char|declare|sitename|net user|xp_cmdshell|or|like'|create|like|table|from|grant|use|group_concat|column_name|information_schema.columns|table_schema|union|where|order|by|\*|or|;|-|\+|,|'|//|/|%|#`)


func validateQuery(query string) bool {
	if len(query) > 100 {
		return false
	}

	if !utf8.Valid([]byte(query)) {
		return false
	}

	letterLower := strings.ToLower(query)
	matched := dangerousPattern.MatchString(letterLower)
	return !matched
}

func chartSQL(baseQuery string, req *request.LogQueryRequest) (string, int64, error) {
	if !util.IsValidIdentifier(req.DataBase) {
		return "", 0, fmt.Errorf("invalid request parameters: %s", req.DataBase)
	}

	if !util.IsValidIdentifier(req.TableName) {
		return "", 0, fmt.Errorf("invalid request parameters: %s", req.TableName)
	}

	if !validateQuery(req.Query) {
		return "", 0, fmt.Errorf("invalid request query: %s", req.Query)
	}
	// Validate request parameters
	if err := validateLogChartRequest(req); err != nil {
		return "", 0, fmt.Errorf("invalid request parameters: %w", err)
	}

	// Calculate interval and get group by clause
	group, interval, err := calculateInterval((req.EndTime-req.StartTime)/1000000, req.TimeField)
	if err != nil {
		return "", 0, fmt.Errorf("failed to calculate interval: %w", err)
	}

	// Build safe query condition
	condition := NewQueryCondition(req.StartTime, req.EndTime, req.TimeField, req.Query)

	// Escape SQL identifiers
	timeField := util.EscapeSQLString(req.TimeField)
	database := util.EscapeSQLString(req.DataBase)
	tableName := util.EscapeSQLString(req.TableName)

	// Build the query
	sql := fmt.Sprintf(baseQuery,
		timeField,
		group,
		database,
		tableName,
		condition,
		group,
		group)
	return sql, interval, nil
}

func (ch *chRepo) GetLogChart(req *request.LogQueryRequest) ([]map[string]any, int64, error) {
	// Build and execute query
	sql, interval, err := chartSQL(queryLogChart, req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build query: %w", err)
	}

	results, err := ch.queryRowsData(sql)
	if err != nil {
		return nil, interval, fmt.Errorf("failed to execute query: %w", err)
	}

	return results, interval, nil
}
