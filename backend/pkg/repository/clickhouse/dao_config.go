// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"log"

	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (ch *chRepo) ModifyTableTTL(ctx core.Context, mapResult []model.ModifyTableTTLMap) error {
	if len(mapResult) == 0 {
		return nil
	}

	for _, table := range mapResult {
		go func(table model.ModifyTableTTLMap) {
			cluster := config.GetCHCluster()
			escapedTableName := fmt.Sprintf("`%s`", table.Name)
			var finalQuery string
			if len(cluster) > 0 {
				escapedClusterName := fmt.Sprintf("`%s`", cluster)
				finalQuery = fmt.Sprintf(`
				ALTER TABLE %s ON CLUSTER %s
				MODIFY %s`,
					escapedTableName, escapedClusterName, table.TTLExpression)
			} else {
				finalQuery = fmt.Sprintf(`
				ALTER TABLE %s
				MODIFY %s`,
					escapedTableName, table.TTLExpression)
			}

			if err := ch.GetContextDB(ctx).Exec(ctx.GetContext(), finalQuery); err != nil {
				log.Printf("failed to modify TTL for table %s: %v\n\n", table.Name, err)
			}
		}(table)
	}

	return nil
}

func (ch *chRepo) GetTables(ctx core.Context, tables []model.Table) ([]model.TablesQuery, error) {
	result := make([]model.TablesQuery, 0)
	query := "SELECT name, create_table_query FROM system.tables WHERE database=(SELECT currentDatabase()) AND name NOT LIKE '.%'"
	var args []interface{}
	argIndex := 1

	names := make([]string, len(tables))
	for i := range tables {
		names = append(names, tables[i].TableName())
	}

	if len(tables) > 0 {
		query += fmt.Sprintf(" AND name IN ($%d)", argIndex)
		args = append(args, names)
		argIndex++
	}

	rows, err := ch.GetContextDB(ctx).Query(ctx.GetContext(), query, args...)
	if err != nil {
		log.Println("Query failed:", err)
		return nil, err
	}
	for rows.Next() {
		row := model.TablesQuery{}
		err := rows.Scan(&row.Name, &row.CreateTableQuery)
		if err != nil {
			log.Println("error to get tables row:", err)
			return result, err
		}
		result = append(result, row)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}
