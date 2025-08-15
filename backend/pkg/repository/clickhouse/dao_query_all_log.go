// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	logsBaseQuery = "SELECT * FROM `%s`.`%s` WHERE %s %s"
)

func (ch *chRepo) QueryAllLogs(ctx core.Context, req *request.LogQueryRequest) ([]map[string]any, string, error) {
	condition := NewQueryCondition(req.StartTime, req.EndTime, req.TimeField, req.Query)
	bySql := NewByLimitBuilder().
		OrderBy(fmt.Sprintf("`%s`", req.TimeField), false).
		Limit(req.PageSize).
		Offset(req.PageNum).
		String()
	sql := buildAllLogsQuery(logsBaseQuery, req, condition, bySql)

	results, err := ch.queryRowsData(ctx, sql)
	if err != nil {
		return nil, sql, err
	}

	return results, sql, nil
}

func (ch *chRepo) QueryAllLogsInOrder(ctx core.Context, req *request.LogQueryRequest) ([]map[string]any, string, error) {
	condition := NewQueryCondition(req.StartTime, req.EndTime, req.TimeField, req.Query)
	bySql := NewByLimitBuilder().
		OrderBy(fmt.Sprintf("`%s`", req.TimeField), true).
		Limit(req.PageSize).
		Offset(0).
		String()
	sql := buildAllLogsQuery(logsBaseQuery, req, condition, bySql)

	results, err := ch.queryRowsData(ctx, sql)
	if err != nil {
		return nil, sql, err
	}

	return results, sql, nil
}

func buildAllLogsQuery(baseQuery string, req *request.LogQueryRequest, condition string, bySql string) string {
	return fmt.Sprintf(baseQuery, req.DataBase, req.TableName, condition, bySql)
}
