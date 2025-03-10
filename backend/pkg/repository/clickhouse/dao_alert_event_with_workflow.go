// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD_COUNT = `SELECT COUNT(DISTINCT alert_id) as count FROM alert_event %s`

const SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD = `WITH paged_alerts AS (
	SELECT *
	FROM (
		SELECT *
		FROM alert_event
		%s
		ORDER BY received_time DESC
		LIMIT 1 BY alert_id
	)
	%s
),
filtered_workflows AS (
    SELECT *
    FROM workflow_records
    %s
)
SELECT
	ae.id,
	ae.group,
    ae.name,
    ae.alert_id,
	ae.create_time,
	ae.update_time,
	ae.end_time,
    ae.received_time,
    ae.detail,
    ae.status,
    ae.tags,
	ae.source,
    toStartOfFiveMinutes(ae.received_time) + INTERVAL 5 MINUTE AS rounded_time,
    wr.workflow_run_id,
    wr.workflow_id,
    wr.workflow_name,
    CASE
      WHEN output = 'false' THEN 'true'
      WHEN output = 'true' THEN 'false'
      ELSE 'unknown'
    END as is_valid
FROM paged_alerts AS ae
LEFT JOIN filtered_workflows AS wr
ON ae.alert_id = wr.ref AND rounded_time = toStartOfFiveMinutes(wr.created_at)
%s`

const SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD_VALID_FIRST = `WITH paged_alerts AS (
	SELECT *
	FROM alert_event
	%s
	ORDER BY received_time DESC
	LIMIT 1 BY alert_id
),
filtered_workflows AS (
    SELECT *,
	CASE
      WHEN output = 'false' THEN 0
      WHEN output = 'true' THEN 1
      ELSE 2
    END as importance
    FROM workflow_records
    %s
)
SELECT
	ae.id,
	ae.group,
    ae.name,
    ae.alert_id,
	ae.create_time,
	ae.update_time,
	ae.end_time,
    ae.received_time,
    ae.detail,
    ae.status,
    ae.tags,
	ae.source,
    toStartOfFiveMinutes(ae.received_time) + INTERVAL 5 MINUTE AS rounded_time,
    wr.workflow_run_id,
    wr.workflow_id,
    wr.workflow_name,
	wr.importance,
    CASE
      WHEN output = 'false' THEN 'true'
      WHEN output = 'true' THEN 'false'
      ELSE 'unknown'
    END as is_valid
FROM paged_alerts AS ae
LEFT JOIN filtered_workflows AS wr
ON ae.alert_id = wr.ref AND rounded_time = toStartOfFiveMinutes(wr.created_at)
%s`

func sortbyParam(sortBy string) ([]string, []bool) {
	if len(sortBy) == 0 {
		return []string{"received_time"}, []bool{true}
	}

	sortBys := strings.Split(sortBy, ",")

	var fields []string
	var ascs []bool
	for _, option := range sortBys {
		parts := strings.Split(option, " ")
		var order string = "asc"
		if len(parts) == 2 {
			order = parts[1]
		}

		if parts[0] == "receivedTime" {
			fields = append(fields, "received_time")
			ascs = append(ascs, order == "asc")
		} else if parts[0] == "isValid" {
			fields = append(fields, "importance")
			ascs = append(ascs, order == "asc")

			if len(sortBys) == 1 {
				fields = append(fields, "received_time")
				ascs = append(ascs, order == "asc")
			}
		}
	}

	return fields, ascs
}

func (ch *chRepo) GetAlertEventWithWorkflowRecord(req *request.AlertEventSearchRequest) ([]alert.AEventWithWRecord, int64, error) {
	alertFilter := NewQueryBuilder().
		Between("received_time", req.StartTime/1e6, req.EndTime/1e6)

	if len(req.Filter.Namespaces) > 0 {
		alertFilter.InStrings("tags['namespace']", req.Filter.Namespaces)
	}
	if len(req.Filter.Nodes) > 0 {
		alertFilter.InStrings("tags['node']", req.Filter.Nodes)
	}

	var count uint64
	countSql := fmt.Sprintf(SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD_COUNT, alertFilter.String())
	err := ch.conn.QueryRow(context.Background(), countSql, alertFilter.values...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	sql, values := getSqlAndValueForSortedAlertEvent(req)

	result := make([]alert.AEventWithWRecord, 0)
	err = ch.conn.Select(context.Background(), &result, sql, values...)
	return result, int64(count), err
}

func getSqlAndValueForSortedAlertEvent(req *request.AlertEventSearchRequest) (string, []any) {
	alertFilter := NewQueryBuilder().
		Between("received_time", req.StartTime/1e6, req.EndTime/1e6)

	if len(req.Filter.Namespaces) > 0 {
		alertFilter.InStrings("tags['namespace']", req.Filter.Namespaces)
	}
	if len(req.Filter.Nodes) > 0 {
		alertFilter.InStrings("tags['node']", req.Filter.Nodes)
	}

	alertOrder := NewByLimitBuilder()
	fields, ascs := sortbyParam(req.SortBy)

	var hasInValid bool
	for idx, field := range fields {
		if field == "importance" {
			hasInValid = true
		}
		alertOrder.OrderBy(field, ascs[idx])
	}
	if req.Pagination != nil {
		alertOrder.Offset((req.Pagination.CurrentPage - 1) * req.Pagination.PageSize).
			Limit(req.Pagination.PageSize)
	}

	intervalMicro := int64(5*time.Minute) / 1e3
	recordFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, (req.EndTime+intervalMicro)/1e6)

	var sql string
	if hasInValid {
		sql = fmt.Sprintf(SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD_VALID_FIRST,
			alertFilter.String(),
			recordFilter.String(),
			alertOrder.String(),
		)
	} else {
		var finalOrder = NewByLimitBuilder()
		for idx, field := range fields {
			finalOrder.OrderBy(field, ascs[idx])
		}
		sql = fmt.Sprintf(SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD,
			alertFilter.String(), alertOrder.String(),
			recordFilter.String(),
			finalOrder.String(),
		)
	}

	values := make([]any, 0)
	values = append(values, alertFilter.values...)
	values = append(values, recordFilter.values...)

	return sql, values
}
