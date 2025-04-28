// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

const SchemaPrefix = "alert_input_schema_"

var (
	AllowSchema = regexp.MustCompile("^[a-zA-Z0-9_-]{1,40}$")
)

func (repo *subRepo) CreateSchema(schema string, columns []string) error {
	if !AllowSchema.MatchString(schema) {
		return alert.ErrNotAllowSchema{Table: schema}
	}

	schema = SchemaPrefix + schema

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
	if !AllowSchema.MatchString(schema) {
		return nil, nil, alert.ErrNotAllowSchema{Table: schema}
	}
	schema = SchemaPrefix + schema

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
func (repo *subRepo) DeleteSchema(schema string) error {
	var enrichRules []alert.AlertEnrichRule

	err := repo.db.Find(&enrichRules, "schema = ?", schema).Error
	if err != nil {
		return err
	}

	var ruleIds []string
	for _, enrichRule := range enrichRules {
		ruleIds = append(ruleIds, enrichRule.EnrichRuleID)
	}

	err = repo.db.Delete(&alert.AlertEnrichSchemaTarget{}, "enrich_rule_id in ?", ruleIds).Error
	if err != nil {
		return err
	}

	schema = SchemaPrefix + schema

	return repo.db.Migrator().DropTable(schema)
}

func (repo *subRepo) ListSchemaColumns(schema string) ([]string, error) {
	if !AllowSchema.MatchString(schema) {
		return nil, alert.ErrNotAllowSchema{Table: schema}
	}
	schema = SchemaPrefix + schema

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
	if !AllowSchema.MatchString(schema) {
		return alert.ErrNotAllowSchema{Table: schema}
	}

	updateTemp, err := buildUpdateSchema(schema, columns)
	if err != nil {
		return err
	}
	for idx, row := range rows {
		var args = make([]interface{}, 0, len(row)+1)
		for _, value := range row {
			args = append(args, value)
		}
		args = append(args, idx)

		err = repo.db.Exec(updateTemp, args...).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func buildUpdateSchema(schema string, columns []string) (string, error) {
	schema = SchemaPrefix + schema

	var columnPlaceHolder []string
	for _, column := range columns {
		if !AllowSchema.MatchString(column) {
			return "", alert.ErrNotAllowSchema{Column: column}
		}

		columnPlaceHolder = append(columnPlaceHolder, fmt.Sprintf("%s = ?", column))
	}

	updateTemp := fmt.Sprintf("UPDATE %s SET %s WHERE schema_row_id = ?", schema, strings.Join(columnPlaceHolder, ","))
	return updateTemp, nil
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
	if !AllowSchema.MatchString(schema) {
		return alert.ErrNotAllowSchema{Table: schema}
	}

	schema = SchemaPrefix + schema
	return repo.clearSchemaData(schema)
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

const insertSchema = "INSERT INTO %s (%s) VALUES %s"

func buildInsertSchema(schema string, columns []string, fullRows [][]string) (string, []interface{}) {
	schema = SchemaPrefix + schema

	colPart := strings.Join(columns, ",")

	valueRows := []string{}
	params := []interface{}{}
	for _, row := range fullRows {
		placeholders := make([]string, len(row))
		for i := range row {
			placeholders[i] = "?"
			params = append(params, row[i])
		}
		valueRows = append(valueRows, fmt.Sprintf("(%s)", strings.Join(placeholders, ",")))
	}

	sql := fmt.Sprintf(insertSchema,
		schema,
		colPart,
		strings.Join(valueRows, ","))

	return sql, params
}

func (repo *subRepo) InsertSchemaData(schema string, columns []string, fullRows [][]string) error {
	if !AllowSchema.MatchString(schema) {
		return alert.ErrNotAllowSchema{Table: schema}
	}
	sql, params := buildInsertSchema(schema, columns, fullRows)

	return repo.db.Raw(sql, params...).Error
}

func (repo *subRepo) clearSchemaData(schema string) error {
	if !AllowSchema.MatchString(schema) {
		return alert.ErrNotAllowSchema{Table: schema}
	}
	sql := "TRUNCATE TABLE " + schema + ";"

	return repo.db.Exec(sql).Error
}
