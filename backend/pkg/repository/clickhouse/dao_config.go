package clickhouse

import (
	"context"
	"fmt"
	"log"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (ch *chRepo) ModifyTableTTL(ctx context.Context, mapResult []model.ModifyTableTTLMap) error {
	if len(mapResult) == 0 {
		return nil
	}

	for _, table := range mapResult {
		go func(table model.ModifyTableTTLMap) {
			escapedTableName := fmt.Sprintf("`%s`", table.Name)
			finalQuery := fmt.Sprintf(`
				ALTER TABLE %s
				MODIFY TTL %s;`,
				escapedTableName, table.TTLExpression)

			if err := ch.conn.Exec(ctx, finalQuery); err != nil {
				fmt.Printf("failed to modify TTL for table %s: %v\n", table.Name, err)
			}
		}(table)
	}

	return nil
}

func (ch *chRepo) GetTables(blackTableNames []string, whiteTableNames []string) ([]model.TablesQuery, error) {
	result := make([]model.TablesQuery, 0)
	query := "SELECT name, create_table_query FROM system.tables WHERE database=(SELECT currentDatabase()) AND name NOT LIKE '.%'"

	args := []interface{}{}
	argIndex := 1

	if len(blackTableNames) > 0 {
		query += fmt.Sprintf(" AND name NOT IN ($%d)", argIndex)
		args = append(args, blackTableNames)
		argIndex++
	}

	if len(whiteTableNames) > 0 {
		query += fmt.Sprintf(" AND name IN ($%d)", argIndex)
		args = append(args, whiteTableNames)
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
