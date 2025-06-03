// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
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

func (ch *chRepo) OtherLogTable(ctx core.Context) ([]map[string]any, error) {
	return ch.queryRowsData(ctx, queryOtherTablesSQL)
}

func (ch *chRepo) OtherLogTableInfo(ctx core.Context, req *request.OtherTableInfoRequest) ([]map[string]any, error) {
	sql := fmt.Sprintf(queryOtherTableInfoSQL, req.DataBase, req.TableName)
	return ch.queryRowsData(ctx, sql)
}
