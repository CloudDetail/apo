// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

// Predefined valid keys for tag filtering
var keys = []string{"source", "container_id", "pid", "container_name", "host_ip", "host_name", "k8s_namespace_name", "k8s_pod_name"}

// isKey checks if the given key is in the predefined list of valid keys
func isKey(key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

// validateTagValue checks if the tag value contains only valid characters
func validateTagValue(value string) bool {
	// Allow letters, numbers, underscores, hyphens, dots, and colons
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
			r == '_' || r == '-' || r == '.' || r == ':') {
			return false
		}
	}
	return true
}

// tagsCondition builds a safe SQL condition for tag filtering
func tagsCondition(tags map[string]string) (string, error) {
	var res string
	for k, v := range tags {
		if !isKey(k) {
			continue
		}
		if !validateTagValue(v) {
			return "", fmt.Errorf("invalid tag value format for key %s", k)
		}
		res += fmt.Sprintf(`AND %s='%s'`, k, util.EscapeSQLString(v))
	}
	if res == "" {
		res = "AND (1='1')"
	}
	return res, nil
}

func reverseSlice(s []map[string]any) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func (ch *chRepo) QueryLogContext(req *request.LogQueryContextRequest) ([]map[string]any, []map[string]any, error) {
	// Validate time parameter
	if req.Time <= 0 {
		return nil, nil, fmt.Errorf("invalid time parameter")
	}

	logtime := req.Time / 1000000

	// Build front time condition
	timefront := fmt.Sprintf("toUnixTimestamp(timestamp) < %d AND toUnixTimestamp(timestamp) > %d ", logtime, logtime-60)

	// Build safe tag conditions
	tags, err := tagsCondition(req.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid tag conditions: %w", err)
	}

	// Build front query
	bySqlfront := NewByLimitBuilder().
		OrderBy("timestamp", false).
		Limit(50).
		String()
	database := util.EscapeSQLString(req.DataBase)
	tablename := util.EscapeSQLString(req.TableName)
	frontSql := fmt.Sprintf(logsBaseQuery, database, tablename, timefront+tags, bySqlfront)
	front, err := ch.queryRowsData(frontSql)
	if err != nil {
		front = []map[string]any{}
	}
	reverseSlice(front)

	// Build end time condition
	timeend := fmt.Sprintf("toUnixTimestamp(timestamp) >= %d AND toUnixTimestamp(timestamp) < %d ", logtime, logtime+60)

	// Build end query
	bySqlend := NewByLimitBuilder().
		OrderBy("timestamp", true).
		Limit(50).
		String()

	endSql := fmt.Sprintf(logsBaseQuery, database, tablename, timeend+tags, bySqlend)
	end, err := ch.queryRowsData(endSql)
	if err != nil {
		end = []map[string]any{}
	}

	return front, end, nil
}
