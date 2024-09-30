package clickhouse

import (
	"context"
	"fmt"
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"log"
)

func (ch *chRepo) ModifyTableTTL(ctx context.Context, mapResult []model.ModifyTableTTLMap) error {
	if len(mapResult) == 0 {
		return nil
	}

	for _, table := range mapResult {
		go func(table model.ModifyTableTTLMap) {
			cluster := config.GetCHCluster()
			escapedTableName := fmt.Sprintf("`%s`", table.Name)
			var finalQuery string
			if len(cluster) > 0 {
				finalQuery = fmt.Sprintf(`
				ALTER TABLE %s ON CLUSTER %s
				MODIFY TTL %s`,
					escapedTableName, cluster, table.TTLExpression)
			} else {
				finalQuery = fmt.Sprintf(`
				ALTER TABLE %s
				MODIFY TTL %s`,
					escapedTableName, table.TTLExpression)
			}

			if err := ch.conn.Exec(ctx, finalQuery); err != nil {
				log.Printf("failed to modify TTL for table %s: %v\n\n", table.Name, err)
			}
		}(table)
	}

	return nil
}

func (ch *chRepo) GetTables(tables []model.Table) ([]model.TablesQuery, error) {
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

	rows, err := ch.conn.Query(context.Background(), query, args...)
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

	// 检查迭代过程中是否有错误
	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}
