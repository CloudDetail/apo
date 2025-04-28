// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"strings"
)

func ValidateSQL(sql string) (string, error) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return "", fmt.Errorf("SQL statement is empty")
	}

	sqlLower := strings.ToLower(sql)

	if strings.Contains(sqlLower, "delete from") {
		if !strings.Contains(sqlLower, "where") {
			return "", fmt.Errorf("DELETE without WHERE clause is not allowed")
		}
	}

	validCommands := []string{"select", "insert", "update", "delete", "create", "alter", "drop", "truncate"}
	hasValidCommand := false
	for _, cmd := range validCommands {
		if strings.Contains(sqlLower, cmd) {
			hasValidCommand = true
			break
		}
	}
	if !hasValidCommand {
		return "", fmt.Errorf("SQL statement lacks a recognized command (e.g., SELECT, INSERT, UPDATE)")
	}

	return sql, nil
}
