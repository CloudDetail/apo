// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
	"time"
)

type ProfilingEvent struct {
	Timestamp       time.Time         `json:"timestamp" ch:"timestamp"`
	StartTime       uint64            `json:"startTime" ch:"startTime"`
	EndTime         uint64            `json:"endTime" ch:"endTime"`
	Offset          int64             `json:"offset" ch:"offset"`
	PID             uint32            `json:"pid" ch:"pid"`
	TID             uint32            `json:"tid" ch:"tid"`
	TransactionIDs  string            `json:"transactionIds" ch:"transactionIds"`
	CPUEvents       string            `json:"cpuEvents" ch:"cpuEvents"`
	InnerCalls      string            `json:"innerCalls" ch:"innerCalls"`
	JavaFutexEvents string            `json:"javaFutexEvents" ch:"javaFutexEvents"`
	Spans           string            `json:"spans" ch:"spans"`
	ThreadName      string            `json:"threadName" ch:"threadName"` // thread name table in labels
	Labels          map[string]string `json:"labels" ch:"labels"`
}

const profiling_event_sql = `SELECT %s FROM profiling_event %s LIMIT %s`

func (ch *chRepo) GetOnOffCPU(pid uint32, nodeName string, startTime, endTime int64) (*[]ProfilingEvent, error) {
	queryBuilder := NewQueryBuilder()
	querySql := queryBuilder.
		Between("startTime", startTime, endTime).
		Between("endTime", startTime, endTime).
		EqualsNotEmpty("labels['node_name']", nodeName).
		Equals("pid", pid).String()
	fieldSql := NewFieldBuilder().
		Fields("timestamp").
		Fields("innerCalls").
		Fields("pid").
		Fields("tid").
		Fields("transactionIds").
		Fields("cpuEvents").
		Fields("javaFutexEvents").
		Fields("labels").
		Alias("intDiv(startTime, 1000)", "startTime").
		Alias("intDiv(endTime, 1000)", "endTime").String()
	sql := fmt.Sprintf(profiling_event_sql, fieldSql, querySql, "10000")
	result := make([]ProfilingEvent, 0)
	err := ch.conn.Select(context.Background(), &result, sql, queryBuilder.values...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
