// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

const GET_ALERTS_WITH_EVENT_COUNT = `WITH event_count AS (
  SELECT alert_id,count(1) as count
  FROM alert_event
  %s
  GROUP BY alert_id
),
alert AS (
  SELECT source,source_id,group,name,tags,alert_id,received_time
  FROM alert_event
  %s
  ORDER BY received_time DESC LIMIT 1 BY alert_id
)
SELECT source,source_id,group,name,tags,alert_id,count
FROM alert a LEFT JOIN event_count ec on a.alert_id = ec.alert_id
ORDER BY received_time DESC LIMIT 5000;
`

func (ch *chRepo) GetAlertsWithEventCount(
	startTime, endTime time.Time,
	filter *alert.AlertEventFilter, maxSize int,
) ([]alert.AlertWithEventCount, uint64, error) {
	whereSQL := extractAlertEventFilter(filter)
	alertFilter := NewQueryBuilder().
		Between("received_time", startTime.Unix(), endTime.Unix()).
		And(whereSQL)

	var count uint64
	countSql := buildAlertQuery(GET_ALERT_EVENTS_COUNT, alertFilter)
	err := ch.conn.QueryRow(context.Background(), countSql, alertFilter.values...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	var values []any
	values = append(values, alertFilter.values...)
	values = append(values, alertFilter.values...)

	sql := fmt.Sprintf(GET_ALERTS_WITH_EVENT_COUNT, alertFilter.String(), alertFilter.String())
	result := make([]alert.AlertWithEventCount, 0)
	err = ch.conn.Select(context.Background(), &result, sql, values...)
	return result, count, err
}

func buildAlertQuery(baseQuery string, queryBuilder *QueryBuilder) string {
	return fmt.Sprintf(baseQuery, queryBuilder.String())
}
