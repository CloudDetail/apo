package clickhouse

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	querySQl = `SELECT * FROM %s.%s %s %s`
)

func (ch *chRepo) queryLogs(sql string) ([]map[string]any, error) {
	rows, err := ch.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		entry := make(map[string]interface{})

		for i, col := range columns {
			entry[col] = values[i]
		}
		results = append(results, entry)
	}
	return results, nil
}

func (ch *chRepo) QueryAllLogs(req *request.LogQueryRequest) ([]map[string]any, string, error) {
	condition := fmt.Sprintf("Where timestamp >= toDateTime64(%d, 3) AND timestamp < toDateTime64(%d, 3) AND %s", req.StartTime/1000000, req.EndTime/1000000, req.Query)
	bySql := NewByLimitBuilder().
		OrderBy("timestamp", false).
		Limit(req.PageSize).
		Offset((req.PageNum - 1) * req.PageSize).
		String()
	sql := fmt.Sprintf(querySQl, req.DataBase, req.TableName, condition, bySql)

	results, err := ch.queryLogs(sql)
	if err != nil {
		return nil, sql, err
	}

	return results, sql, nil
}
