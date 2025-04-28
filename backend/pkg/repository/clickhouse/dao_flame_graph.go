// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
)

type FlameGraphData struct {
	StartTime   int64             `ch:"start_time" json:"startTime"`
	EndTime     int64             `ch:"end_time" json:"endTime"`
	PID         uint32            `ch:"pid" json:"pid"`
	TID         uint32            `ch:"tid" json:"tid"`
	SampleType  string            `ch:"sample_type" json:"sampleType"`
	SampleRate  uint32            `ch:"sample_rate" json:"sampleRate"`
	Labels      map[string]string `ch:"labels" json:"labels"`
	FlameBearer string            `ch:"flamebearer" json:"flameBearer"`
}

const flame_graph_sql = `SELECT DISTINCT toUnixTimestamp64Nano(start_time) as start_time, toUnixTimestamp64Nano(end_time) as end_time, pid, tid, sample_type, sample_rate, labels, flamebearer FROM flame_graph %s ORDER BY start_time DESC`

func (ch *chRepo) GetFlameGraphData(startTime, endTime int64, nodeName string, pid, tid int64, sampleType, spanId, traceId string) (*[]FlameGraphData, error) {
	queryBuilder := NewQueryBuilder()
	queryBuilder.Between("start_time", startTime*1000, endTime*1000).
		Between("end_time", startTime*1000, endTime*1000).
		EqualsNotEmpty("sample_type", sampleType).
		EqualsNotEmpty("labels['span_id']", spanId).
		EqualsNotEmpty("labels['trace_id']", traceId).
		EqualsNotEmpty("labels['node_name']", nodeName)
	if pid > 0 {
		queryBuilder.Equals("pid", pid)
	}
	if tid >= 0 {
		queryBuilder.Equals("tid", tid)
	}
	sql := buildFlameGraphQuery(queryBuilder)
	result := make([]FlameGraphData, 0)
	err := ch.conn.Select(context.Background(), &result, sql, queryBuilder.values...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func buildFlameGraphQuery(queryBuilder *QueryBuilder) string {
	sql := fmt.Sprintf(flame_graph_sql, queryBuilder.String())
	return sql
}
