// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	core "github.com/CloudDetail/apo/backend/pkg/core"
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

func (ch *chRepo) OtherLogTable(ctx_core core.Context,) ([]map[string]any, error) {
	return ch.queryRowsData(queryOtherTablesSQL)
}

func (ch *chRepo) OtherLogTableInfo(ctx_core core.Context, req *request.OtherTableInfoRequest) ([]map[string]any, error) {
	sql := fmt.Sprintf(queryOtherTableInfoSQL, req.DataBase, req.TableName)
	return ch.queryRowsData(sql)
}
