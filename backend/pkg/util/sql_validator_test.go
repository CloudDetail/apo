// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"testing"
)

func TestValidateSQL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedSQL string
		expectedErr error
	}{
		{
			name:        "Empty SQL",
			input:       "",
			expectedSQL: "",
			expectedErr: fmt.Errorf("SQL statement is empty"),
		},
		{
			name:        "Whitespace SQL",
			input:       "   ",
			expectedSQL: "",
			expectedErr: fmt.Errorf("SQL statement is empty"),
		},
		// pass
		{
			name:        "DROP TABLE statement",
			input:       "DROP TABLE users;",
			expectedSQL: "DROP TABLE users;",
			expectedErr: nil,
		},
		// pass
		{
			name:        "DROP DATABASE statement",
			input:       "DROP DATABASE test;",
			expectedSQL: "DROP DATABASE test;",
			expectedErr: nil,
		},
		// pass
		{
			name:        "TRUNCATE statement",
			input:       "TRUNCATE TABLE users;",
			expectedSQL: "TRUNCATE TABLE users;",
			expectedErr: nil,
		},
		{
			name:        "DELETE without WHERE",
			input:       "DELETE FROM users;",
			expectedSQL: "",
			expectedErr: fmt.Errorf("DELETE without WHERE clause is not allowed"),
		},
		{
			name:        "Valid SELECT",
			input:       "SELECT * FROM users;",
			expectedSQL: "SELECT * FROM users;",
			expectedErr: nil,
		},
		{
			name:        "Valid INSERT",
			input:       "INSERT INTO users (name) VALUES ('John');",
			expectedSQL: "INSERT INTO users (name) VALUES ('John');",
			expectedErr: nil,
		},
		{
			name:        "Valid UPDATE",
			input:       "UPDATE quick_alert_rule_metric SET name_en = 'Average Request Latency' WHERE name = '平均请求延时' AND name_en IS NULL;",
			expectedSQL: "UPDATE quick_alert_rule_metric SET name_en = 'Average Request Latency' WHERE name = '平均请求延时' AND name_en IS NULL;",
			expectedErr: nil,
		},
		{
			name:        "Valid DELETE with WHERE",
			input:       "DELETE FROM users WHERE id = 1;",
			expectedSQL: "DELETE FROM users WHERE id = 1;",
			expectedErr: nil,
		},
		{
			name:        "Valid CREATE",
			input:       "CREATE TABLE users (id INT);",
			expectedSQL: "CREATE TABLE users (id INT);",
			expectedErr: nil,
		},
		{
			name:        "Valid ALTER",
			input:       "ALTER TABLE users ADD COLUMN name VARCHAR(50);",
			expectedSQL: "ALTER TABLE users ADD COLUMN name VARCHAR(50);",
			expectedErr: nil,
		},
		{
			name:        "Case insensitive SQL",
			input:       "select * from users;",
			expectedSQL: "select * from users;",
			expectedErr: nil,
		},
		{
			name:        "SQL with leading/trailing spaces",
			input:       "  SELECT * FROM users;  ",
			expectedSQL: "SELECT * FROM users;",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSQL(tt.input)
			if result != tt.expectedSQL {
				t.Errorf("Expected SQL %q, got %q", tt.expectedSQL, result)
			}
			if err != nil && tt.expectedErr != nil {
				if err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %q, got %q", tt.expectedErr, err)
				}
			} else if err != tt.expectedErr {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
