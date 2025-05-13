// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
	"time"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

type NetSegments struct {
	StartTime		time.Time	`ch:"start_time"`
	EndTime			time.Time	`ch:"end_time"`
	ResponseDuration	uint64		`ch:"response_duration"`
	TapSide			string		`ch:"tap_side"`
	SpanId			string		`ch:"span_id"`
	TraceId			string		`ch:"trace_id"`
}

func (ch *chRepo) GetNetworkSpanSegments(ctx_core core.Context, traceId string, spanId string) ([]NetSegments, error) {
	queryBuilder := NewQueryBuilder().
		EqualsNotEmpty("trace_id", traceId).
		EqualsNotEmpty("span_id", spanId)
	executeSql := buildQuery(queryBuilder)
	var netSegments []NetSegments
	if err := ch.conn.Select(context.Background(), &netSegments, executeSql, queryBuilder.values...); err != nil {
		return nil, err
	}
	return netSegments, nil
}

func buildQuery(queryBuilder *QueryBuilder) string {
	spanSegmentSqlTemplate := "SELECT %s FROM flow_log.l7_flow_log %s"
	fields := "start_time, end_time, response_duration, tap_side, span_id, trace_id"
	executeSql := fmt.Sprintf(spanSegmentSqlTemplate, fields, queryBuilder.String())
	return executeSql
}
