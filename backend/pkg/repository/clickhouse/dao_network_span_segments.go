// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"time"
)

type NetSegments struct {
	StartTime        time.Time `ch:"start_time"`
	EndTime          time.Time `ch:"end_time"`
	ResponseDuration uint64    `ch:"response_duration"`
	TapSide          string    `ch:"tap_side"`
	SpanId           string    `ch:"span_id"`
	TraceId          string    `ch:"trace_id"`
}

func (ch *chRepo) GetNetworkSpanSegments(traceId string, spanId string) ([]NetSegments, error) {
	queryBuilder := NewQueryBuilder().
		EqualsNotEmpty("trace_id", traceId).
		EqualsNotEmpty("span_id", spanId)

	queryBuilder.baseQuery = "SELECT start_time, end_time, response_duration, tap_side, span_id, trace_id FROM flow_log.l7_flow_log "
	var netSegments []NetSegments
	if err := ch.conn.Select(context.Background(), &netSegments, queryBuilder.String(), queryBuilder.values...); err != nil {
		return nil, err
	}
	return netSegments, nil
}
