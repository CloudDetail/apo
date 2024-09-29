package clickhouse

import (
	"context"
	"fmt"
	"github.com/CloudDetail/apo/backend/config"
	"log"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (ch *chRepo) ModifyTableTTL(ctx context.Context, mapResult []model.ModifyTableTTLMap) error {
	if len(mapResult) == 0 {
		return nil
	}

	for _, table := range mapResult {
		go func(table model.ModifyTableTTLMap) {
			cluster := getCluster()
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

func getCluster() string {
	cfg := config.Get()
	return cfg.ClickHouse.Cluster
}

func (ch *chRepo) GetTables(tableNames []string) ([]model.TablesQuery, error) {
	result := make([]model.TablesQuery, 0)
	if len(getCluster()) > 0 {
		for i := range tableNames {
			if !strings.HasSuffix(tableNames[i], "_local") {
				tableNames[i] += "_local"
			}
		}
	}
	query := "SELECT name, create_table_query FROM system.tables WHERE database=(SELECT currentDatabase()) AND name NOT LIKE '.%'"

	args := []interface{}{}
	argIndex := 1

	if len(tableNames) > 0 {
		query += fmt.Sprintf(" AND name IN ($%d)", argIndex)
		args = append(args, tableNames)
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
