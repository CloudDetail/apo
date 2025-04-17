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

const SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD_COUNT = `WITH lastEvent AS (
  SELECT alert_id,status,%s as rounded_time
  FROM alert_event ae
  %s
  ORDER BY received_time DESC LIMIT 1 BY alert_id
),
filtered_workflows AS (
  SELECT rounded_time,ref,output,
  CASE
    WHEN output = 'false' THEN 2
    WHEN output = 'true' THEN 1
    ELSE 0
  END as importance
  FROM workflow_records
  %s
)
SELECT count(1) FROM
  (
    SELECT ae.alert_id,ae.status,
      CASE
        WHEN fw.importance = 0 and fw.output != '' THEN 'failed'
        WHEN fw.importance = 0 and ae.status = 'firing'  THEN 'unknown'
        WHEN fw.importance = 0 and ae.status = 'resolved' THEN 'skipped'
        WHEN fw.importance = 1 THEN 'invalid'
        WHEN fw.importance = 2 THEN 'valid'
      END as validity
    FROM lastEvent ae
    LEFT JOIN filtered_workflows fw on ae.rounded_time = fw.rounded_time and ae.alert_id = fw.ref
   )
%s
`

const SQL_GET_ALERTEVENT_COUNTS = `WITH lastEvent AS (
  SELECT *, %s as rounded_time
  FROM alert_event ae
  %s
  ORDER BY received_time DESC LIMIT 1 BY alert_id
),
filtered_workflows AS (
  SELECT *,
  CASE
    WHEN output = 'false' THEN 2
    WHEN output = 'true' THEN 1
    ELSE 0
  END as importance
  FROM workflow_records
  %s
)
SELECT count(1) as count,validity,status FROM(
  SELECT status,alert_id,
    CASE
      WHEN fw.importance = 0 and fw.output != '' THEN 'failed'
      WHEN fw.importance = 0 and ae.status = 'firing'  THEN 'unknown'
      WHEN fw.importance = 0 and ae.status = 'resolved' THEN 'skipped'
      WHEN fw.importance = 1 THEN 'invalid'
      WHEN fw.importance = 2 THEN 'valid'
    END as validity
  FROM lastEvent ae
  LEFT JOIN filtered_workflows fw on ae.rounded_time = fw.rounded_time and ae.alert_id = fw.ref
) GROUP BY validity,status
`

const SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD = `WITH lastEvent AS (
  SELECT *, %s as rounded_time
  FROM alert_event ae
  %s
  ORDER BY received_time DESC LIMIT 1 BY alert_id
),
filtered_workflows AS (
  SELECT *,
  CASE
    WHEN output = 'false' THEN 2
    WHEN output = 'true' THEN 1
    ELSE 0
  END as importance,
  %s as rounded_time_e
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
  fw.workflow_run_id,
  fw.workflow_id,
  fw.workflow_name,
  fw.importance,
  fw.output,
  CASE
    WHEN output = 'false' THEN 'true'
    WHEN output = 'true' THEN 'false'
    ELSE 'unknown'
  END as is_valid,
  CASE
    WHEN fw.importance = 0 and fw.output != '' THEN 'failed'
    WHEN fw.importance = 0 and ae.status = 'firing'  THEN 'unknown'
    WHEN fw.importance = 0 and ae.status = 'resolved' THEN 'skipped'
    WHEN fw.importance = 1 THEN 'invalid'
    WHEN fw.importance = 2 THEN 'valid'
  END as validity
FROM lastEvent ae
LEFT JOIN filtered_workflows fw
ON ae.alert_id = fw.ref AND ae.rounded_time = fw.rounded_time_e
%s %s`

// Deprecated: Used to be compatible with version 1.5, will be removed after version 1.7.
func getWorkflowRecordRoundedTime(cacheMinutes int) string {
	return fmt.Sprintf(`CASE
      WHEN rounded_time > 0 THEN rounded_time
      ELSE toStartOfInterval(created_at, INTERVAL %d MINUTE)
    END`, cacheMinutes)
}

func getEventRoundedTime(cacheMinutes int) string {
	return fmt.Sprintf(`toStartOfInterval(ae.update_time, INTERVAL %d MINUTE) + INTERVAL %d MINUTE`, cacheMinutes, cacheMinutes)
}

func sortbyParam(sortBy string) ([]string, []bool) {
	if len(sortBy) == 0 {
		return []string{"importance", "update_time"}, []bool{false, true}
	}

	sortBys := strings.Split(sortBy, ",")

	var fields []string
	var ascs []bool
	for _, option := range sortBys {
		parts := strings.Split(option, " ")
		var order string = "desc"
		if len(parts) == 2 {
			order = parts[1]
		}

		if parts[0] == "receivedTime" {
			fields = append(fields, "update_time")
			ascs = append(ascs, order == "asc")
		} else if parts[0] == "isValid" {
			fields = append(fields, "importance")
			ascs = append(ascs, order == "asc")

			if len(sortBys) == 1 {
				fields = append(fields, "update_time")
				ascs = append(ascs, order == "asc")
			}
		}
	}

	return fields, ascs
}

func (ch *chRepo) GetAlertEventWithWorkflowRecord(req *request.AlertEventSearchRequest, cacheMinutes int) ([]alert.AEventWithWRecord, int64, error) {
	alertFilter := NewQueryBuilder().
		Between("update_time", req.StartTime/1e6, req.EndTime/1e6)

	if len(req.Filter.Namespaces) > 0 {
		alertFilter.InStrings("tags['namespace']", req.Filter.Namespaces)
	}
	if len(req.Filter.Nodes) > 0 {
		alertFilter.InStrings("tags['node']", req.Filter.Nodes)
	}

	var count uint64
	intervalMicro := int64(5*time.Minute) / 1e3
	recordFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, (req.EndTime+intervalMicro)/1e6)

	resultFilter := NewQueryBuilder()
	if len(req.Filter.Validity) > 0 {
		resultFilter.InStrings("validity", req.Filter.Validity)
	}
	if len(req.Filter.Status) > 0 {
		resultFilter.InStrings("status", req.Filter.Status)
	}

	countSql := fmt.Sprintf(SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD_COUNT,
		getEventRoundedTime(cacheMinutes),
		alertFilter.String(),
		recordFilter.String(),
		resultFilter.String(),
	)

	values := append(alertFilter.values, recordFilter.values...)
	values = append(values, resultFilter.values...)
	err := ch.conn.QueryRow(context.Background(), countSql, values...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	sql, values := getSqlAndValueForSortedAlertEvent(req, cacheMinutes)

	result := make([]alert.AEventWithWRecord, 0)
	err = ch.conn.Select(context.Background(), &result, sql, values...)
	return result, int64(count), err
}

func getSqlAndValueForSortedAlertEvent(req *request.AlertEventSearchRequest, cacheMinutes int) (string, []any) {
	alertFilter := NewQueryBuilder().
		Between("update_time", req.StartTime/1e6, req.EndTime/1e6)

	if len(req.Filter.Namespaces) > 0 {
		alertFilter.InStrings("tags['namespace']", req.Filter.Namespaces)
	}
	if len(req.Filter.Nodes) > 0 {
		alertFilter.InStrings("tags['node']", req.Filter.Nodes)
	}

	resultOrder := NewByLimitBuilder()
	fields, ascs := sortbyParam(req.SortBy)

	for idx, field := range fields {
		resultOrder.OrderBy(field, ascs[idx])
	}

	if req.Pagination != nil {
		resultOrder.Offset((req.Pagination.CurrentPage - 1) * req.Pagination.PageSize).
			Limit(req.Pagination.PageSize)
	}

	intervalMicro := int64(5*time.Minute) / 1e3
	recordFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, (req.EndTime+intervalMicro)/1e6)

	resultFilter := NewQueryBuilder()
	if len(req.Filter.Validity) > 0 {
		resultFilter.InStrings("validity", req.Filter.Validity)
	}
	if len(req.Filter.Status) > 0 {
		resultFilter.InStrings("status", req.Filter.Status)
	}

	sql := fmt.Sprintf(SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD,
		getEventRoundedTime(cacheMinutes),
		alertFilter.String(),
		getWorkflowRecordRoundedTime(cacheMinutes),
		recordFilter.String(),
		resultFilter.String(),
		resultOrder.String(),
	)

	values := make([]any, 0)
	values = append(values, alertFilter.values...)
	values = append(values, recordFilter.values...)
	values = append(values, resultFilter.values...)

	return sql, values
}

func (ch *chRepo) GetAlertEventCounts(req *request.AlertEventSearchRequest, cacheMinutes int) (map[string]int64, error) {
	alertFilter := NewQueryBuilder().
		Between("update_time", req.StartTime/1e6, req.EndTime/1e6)

	var counts []_alertEventCount
	intervalMicro := int64(5*time.Minute) / 1e3
	recordFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, (req.EndTime+intervalMicro)/1e6)
	countSql := fmt.Sprintf(SQL_GET_ALERTEVENT_COUNTS,
		getEventRoundedTime(cacheMinutes),
		alertFilter.String(),
		recordFilter.String(),
	)
	values := append(alertFilter.values, recordFilter.values...)
	err := ch.conn.Select(context.Background(), &counts, countSql, values...)
	if err != nil {
		return nil, err
	}

	result := map[string]int64{
		"firing":   0,
		"resolved": 0,
		"valid":    0,
		"invalid":  0,
		"skipped":  0,
		"failed":   0,
		"unknown":  0,
	}

	for _, count := range counts {
		result[count.Status] += int64(count.Count)
		result[count.Validity] += int64(count.Count)
	}
	return result, nil
}

type _alertEventCount struct {
	Validity string `ch:"validity"`
	Status   string `ch:"status"`
	Count    uint64 `ch:"count"`
}
