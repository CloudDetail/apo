package clickhouse

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const queryOtherTablesSQL = `
SELECT
    database,
		name,
FROM system.tables;
`

const queryOtherTableInfoSQL = `
SELECT 
    name,type
FROM 
    system.columns
WHERE database = '%s' And table = '%s';
`

func (ch *chRepo) OtherLogTable() ([]map[string]any, error) {
	return ch.queryRowsData(queryOtherTablesSQL)
}

func (ch *chRepo) OtherLogTableInfo(req *request.OtherTableInfoRequest) ([]map[string]any, error) {
	sql := fmt.Sprintf(queryOtherTableInfoSQL, req.DataBase, req.TableName)
	return ch.queryRowsData(sql)
}
