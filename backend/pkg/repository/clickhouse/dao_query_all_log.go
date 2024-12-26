// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	querySQl = "SELECT * FROM `%s`.`%s` WHERE %s %s"
)

func (ch *chRepo) QueryAllLogs(req *request.LogQueryRequest) ([]map[string]any, string, error) {
	condition := NewQueryCondition(req.TimeField, req.Query)
	bySql := NewByLimitBuilder().
		OrderBy(fmt.Sprintf("`%s`", req.TimeField), false).
		Limit(req.PageSize).
		Offset((req.PageNum - 1) * req.PageSize).
		String()
	query := fmt.Sprintf(querySQl, req.DataBase, req.TableName, condition, bySql)

	results, err := ch.queryRowsData(query, req.StartTime/1000000, req.EndTime/1000000)
	if err != nil {
		return nil, query, err
	}

	return results, query, nil
}
