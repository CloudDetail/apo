// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func calculateInterval(interval int64, timeField string) (string, int64) {
	if interval == 0 {
		return "", 0
	}
	if interval <= 60*5 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 1 second)", timeField), 1
	} else if interval <= 60*30 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 1 minute)", timeField), 60
	} else if interval <= 60*60*4 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 10 minute)", timeField), 600
	} else if interval <= 60*60*24 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 1 hour)", timeField), 3600
	} else if interval <= 60*60*24*7 {
		return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 6 hour)", timeField), 21600
	}
	return fmt.Sprintf("toStartOfInterval(`%s`, INTERVAL 1 day)", timeField), 86400
}

const queryLogChart = "SELECT count(`%s`) as count, %s as timeline FROM `%s`.`%s` WHERE %s GROUP BY %s ORDER BY %s ASC"

func chartSQL(baseQuery string, req *request.LogQueryRequest) (string, int64) {
	group, interval := calculateInterval((req.EndTime-req.StartTime)/1000000, req.TimeField)
	condition := NewQueryCondition(req.StartTime, req.EndTime, req.TimeField, req.Query)
	sql := fmt.Sprintf(baseQuery,
		req.TimeField,
		group,
		req.DataBase,
		req.TableName,
		condition,
		group,
		group)
	return sql, interval
}

func (ch *chRepo) GetLogChart(ctx core.Context, req *request.LogQueryRequest) ([]map[string]any, int64, error) {
	sql, interval := chartSQL(queryLogChart, req)
	results, err := ch.queryRowsData(ctx, sql)
	if err != nil {
		return nil, interval, err
	}
	return results, interval, nil
}
