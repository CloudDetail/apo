// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

const SchemaPrefix = "alert_input_schema_"

var (
	AllowSchema = regexp.MustCompile("^[a-zA-Z0-9_-]{1,40}$")
)

func (repo *subRepo) CreateSchema(schema string, columns []string) error {
	schema = SchemaPrefix + schema
	if !AllowSchema.MatchString(schema) {
		return alert.ErrNotAllowSchema{Table: schema}
	}

	var fields []string
	for _, col := range columns {
		if !AllowSchema.MatchString(col) {
			return alert.ErrNotAllowSchema{Table: schema}
		}

		fields = append(fields, fmt.Sprintf("%s VARCHAR(255)", col))
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (schema_row_id integer PRIMARY KEY AUTOINCREMENT,%s);", schema, strings.Join(fields, ", "))

	validSql, err := util.ValidateSQL(sql)
	if err != nil {
		return err
	}

	return repo.db.Exec(validSql).Error
}

func (repo *subRepo) GetSchemaData(schema string) ([]string, map[int64][]string, error) {
	schema = SchemaPrefix + schema
	if !AllowSchema.MatchString(schema) {
		return nil, nil, alert.ErrNotAllowSchema{Table: schema}
	}

	rows, err := repo.db.Table(schema).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()

	var idIdx = 0
	for idx, column := range columns {
		if column == "schema_row_id" {
			idIdx = idx
		}
	}

	entry := make(map[int64][]string)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, err
		}

		var rowId int64 = 0
		var strValues []string
		for i, value := range values {
			if i == idIdx {
				rowId = values[i].(int64)
				continue
			}
			strValues = append(strValues, value.(string))
		}
		entry[rowId] = strValues
	}

	return append(columns[:idIdx], columns[idIdx+1:]...), entry, nil
}

// Delete schema and related alertRules
func (repo *subRepo) DeleteSchema(ctx core.Context, schema string) error {
	var enrichRules []alert.AlertEnrichRule

	err := repo.UserByContext(ctx).Find(&enrichRules, "schema = ?", schema).Error
	if err != nil {
		return err
	}

	var ruleIds []string
	for _, enrichRule := range enrichRules {
		ruleIds = append(ruleIds, enrichRule.EnrichRuleID)
	}

	err = repo.UserByContext(ctx).Delete(&alert.AlertEnrichSchemaTarget{}, "enrich_rule_id in ?", ruleIds).Error
	if err != nil {
		return err
	}

	schema = SchemaPrefix + schema

	return repo.UserByContext(ctx).Migrator().DropTable(schema)
}

func (repo *subRepo) ListSchemaColumns(schema string) ([]string, error) {
	schema = SchemaPrefix + schema
	if !AllowSchema.MatchString(schema) {
		return nil, alert.ErrNotAllowSchema{Table: schema}
	}

	rows, err := repo.db.Table(schema).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()

	var idIdx = 0
	for idx, column := range columns {
		if column == "schema_row_id" {
			idIdx = idx
		}
	}

	return append(columns[:idIdx], columns[idIdx+1:]...), nil
}

func (repo *subRepo) UpdateSchemaData(schema string, columns []string, rows map[int][]string) error {
	schema = SchemaPrefix + schema
	if !AllowSchema.MatchString(schema) {
		return alert.ErrNotAllowSchema{Table: schema}
	}

	for _, column := range columns {
		if !AllowSchema.MatchString(column) {
			return alert.ErrNotAllowSchema{Column: column}
		}
	}

	for idx, values := range rows {
		if len(values) != len(columns) {
			return fmt.Errorf("not allow row: %d", idx)
		}

		for _, value := range values {
			if !util.IsValidIdentifier(value) {
				return fmt.Errorf("invalid row value: %s", value)
			}
		}

		var updateColumns = map[string]any{}
		for i := 0; i < len(columns); i++ {
			updateColumns[columns[i]] = values[i]
		}

		err := repo.db.Table(schema).Where("schema_row_id = ?", idx).Updates(updateColumns).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *subRepo) ListSchema() ([]string, error) {
	tables, err := repo.db.Migrator().GetTables()
	if err != nil {
		return nil, err
	}

	var schemas []string
	for _, table := range tables {
		if strings.HasPrefix(table, SchemaPrefix) {
			schemas = append(schemas, strings.TrimPrefix(table, SchemaPrefix))
		}
	}

	return schemas, nil
}

func (repo *subRepo) ClearSchemaData(schema string) error {
	schema = SchemaPrefix + schema
	if !AllowSchema.MatchString(schema) {
		return alert.ErrNotAllowSchema{Table: schema}
	}

	sql := "TRUNCATE TABLE " + schema + ";"
	return repo.db.Exec(sql).Error
}

func EscapeString(input string) string {
	var builder strings.Builder
	for _, c := range input {
		switch c {
		case 0:
			builder.WriteString("\\0")
		case '\n':
			builder.WriteString("\\n")
		case '\r':
			builder.WriteString("\\r")
		case '\\':
			builder.WriteString("\\\\")
		case '\'':
			builder.WriteString("\\'")
		case '"':
			builder.WriteString("\\\"")
		case '\x1A':
			builder.WriteString("\\Z")
		default:
			builder.WriteRune(c)
		}
	}
	return builder.String()
}

func (repo *subRepo) InsertSchemaData(schema string, columns []string, fullRows [][]string) error {
	schema = SchemaPrefix + schema
	if !AllowSchema.MatchString(schema) {
		return alert.ErrNotAllowSchema{Table: schema}
	}
	//	sql, params := buildInsertSchema(schema, columns, fullRows)

	for _, column := range columns {
		if !AllowSchema.MatchString(column) {
			return alert.ErrNotAllowSchema{Column: column}
		}
	}

	for _, fullRow := range fullRows {
		if len(fullRow) != len(columns) {
			return fmt.Errorf("not allow row: %v", fullRow)
		}

		for _, value := range fullRow {
			if !util.IsValidIdentifier(value) {
				return fmt.Errorf("invalid row value: %s", value)
			}
		}

		var values = map[string]any{}

		for i := 0; i < len(columns); i++ {
			values[columns[i]] = fullRow[i]
		}

		err := repo.db.Table(schema).Create(values).Error
		if err != nil {
			return err
		}
	}

	return nil
}
