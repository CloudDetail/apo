// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
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
WHERE database = ? And table = ?;
`

func (ch *chRepo) OtherLogTable() ([]map[string]any, error) {
	return ch.queryRowsData(queryOtherTablesSQL)
}

func (ch *chRepo) OtherLogTableInfo(req *request.OtherTableInfoRequest) ([]map[string]any, error) {
	return ch.queryRowsData(queryOtherTableInfoSQL, req.DataBase, req.TableName)
}
