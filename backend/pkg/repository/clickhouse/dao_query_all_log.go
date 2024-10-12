package clickhouse

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	querySQl = `SELECT * FROM %s.%s WHERE %s %s`
)

func (ch *chRepo) QueryAllLogs(req *request.LogQueryRequest) ([]map[string]any, string, error) {
	condition := NewQueryCondition(req.StartTime, req.EndTime, req.Query)
	bySql := NewByLimitBuilder().
		OrderBy("timestamp", false).
		Limit(req.PageSize).
		Offset((req.PageNum - 1) * req.PageSize).
		String()
	sql := fmt.Sprintf(querySQl, req.DataBase, req.TableName, condition, bySql)

	results, err := ch.queryRowsData(sql)
	if err != nil {
		return nil, sql, err
	}

	return results, sql, nil
}
