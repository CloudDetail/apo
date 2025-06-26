// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD_COUNT = `WITH lastEvent AS (
  SELECT alert_id,status
  FROM alert_event ae
  %s
  ORDER BY received_time DESC LIMIT 1 BY alert_id
),
filtered_workflows AS (
  SELECT ref,output,
  CASE
    WHEN output = 'false' THEN 2
    WHEN output = 'true' THEN 1
    ELSE 0
  END as importance
  FROM workflow_records
  %s
  ORDER BY created_at DESC LIMIT 1 BY ref
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
    LEFT JOIN filtered_workflows fw on ae.alert_id = fw.ref
   )
%s
`

const SQL_GET_ALERTEVENT_COUNTS = `WITH lastEvent AS (
  SELECT *
  FROM alert_event ae
  %s
  ORDER BY received_time DESC LIMIT 1 BY alert_id
),
filtered_workflows AS (
  SELECT *,
  CASE
    WHEN output = 'false' THEN 3
	WHEN output != '' and output != 'true' THEN 2
    WHEN output = 'true' THEN 1
    ELSE 0
  END as importance
  FROM workflow_records
  %s
  ORDER BY created_at DESC LIMIT 1 BY ref
)
SELECT count(1) as count,validity,status FROM(
  SELECT status,alert_id,
    CASE
      WHEN fw.importance = 3 THEN 'valid'
      WHEN fw.importance = 2 and fw.output != '' THEN 'failed'
      WHEN fw.importance = 1 THEN 'invalid'
      WHEN fw.importance = 0 and ae.status = 'firing'  THEN 'unknown'
      WHEN fw.importance = 0 and ae.status = 'resolved' THEN 'skipped'
    END as validity
  FROM lastEvent ae
  LEFT JOIN filtered_workflows fw on ae.alert_id = fw.ref
) GROUP BY validity,status
`

const SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD = `WITH lastEvent AS (
  SELECT *
  FROM alert_event ae
  %s
  ORDER BY received_time DESC LIMIT 1 BY alert_id
),
filtered_workflows AS (
  SELECT *,
  CASE
    WHEN output = 'false' THEN 3
	WHEN output != '' and output != 'true' THEN 2
    WHEN output = 'true' THEN 1
    ELSE 0
  END as importance
  FROM workflow_records
  %s
  ORDER BY created_at DESC LIMIT 1 BY ref
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
  ae.severity,
  ae.detail,
  ae.status,
  ae.tags,
  ae.raw_tags,
  ae.source,
  fw.workflow_run_id,
  fw.workflow_id,
  fw.workflow_name,
  fw.importance,
  fw.created_at as last_check_at,
  fw.output,
  CASE
    WHEN output = 'false' THEN 'true'
    WHEN output = 'true' THEN 'false'
    ELSE 'unknown'
  END as is_valid,
  CASE
    WHEN fw.importance = 3 THEN 'valid'
    WHEN fw.importance = 2 and fw.output != '' THEN 'failed'
    WHEN fw.importance = 1 THEN 'invalid'
    WHEN fw.importance = 0 and ae.status = 'firing'  THEN 'unknown'
    WHEN fw.importance = 0 and ae.status = 'resolved' THEN 'skipped'
  END as validity
FROM lastEvent ae
LEFT JOIN filtered_workflows fw
ON ae.alert_id = fw.ref
%s %s`

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

func (ch *chRepo) GetAlertEventWithWorkflowRecord(ctx core.Context, req *request.AlertEventSearchRequest, cacheMinutes int) ([]alert.AEventWithWRecord, int64, error) {
	alertFilter := NewQueryBuilder().
		Between("update_time", req.StartTime/1e6, req.EndTime/1e6).
		NotGreaterThan("end_time", req.EndTime/1e6)

	// TODO remove in v1.9.x
	{
		if len(req.Filter.Namespaces) > 0 {
			alertFilter.InStrings("tags['namespace']", req.Filter.Namespaces)
		}
		if len(req.Filter.Nodes) > 0 {
			alertFilter.InStrings("tags['node']", req.Filter.Nodes)
		}
	}

	var count uint64
	intervalMicro := int64(cacheMinutes) * int64(time.Minute) / 1e3
	endTime := req.EndTime/1e6 + int64(5*time.Minute)/1e9
	recordFilter := NewQueryBuilder().Between("created_at", (req.StartTime-intervalMicro)/1e6, endTime)

	resultFilter := NewQueryBuilder()

	// TODO remove in v1.9.x
	{
		if len(req.Filter.Validity) > 0 {
			resultFilter.InStrings("validity", req.Filter.Validity)
		}
		if len(req.Filter.Status) > 0 {
			resultFilter.InStrings("status", req.Filter.Status)
		}
	}

	err := applyFilter(req.Filters, resultFilter, alertFilter)
	if err != nil {
		return nil, 0, err
	}

	countSql := buildCountQuery(alertFilter, recordFilter, resultFilter, cacheMinutes)

	values := append(alertFilter.values, recordFilter.values...)
	values = append(values, resultFilter.values...)
	err = ch.GetContextDB(ctx).QueryRow(ctx.GetContext(), countSql, values...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	sql, values, err := getSqlAndValueForSortedAlertEvent(req, cacheMinutes)
	if err != nil {
		return nil, 0, err
	}

	result := make([]alert.AEventWithWRecord, 0)
	err = ch.GetContextDB(ctx).Select(ctx.GetContext(), &result, sql, values...)

	for i := 0; i < len(result); i++ {
		// result[i].EnrichTagsDisplay =
	}
	return result, int64(count), err
}

func buildCountQuery(alertFilter *QueryBuilder, recordFilter *QueryBuilder, resultFilter *QueryBuilder, cacheMinutes int) string {
	return fmt.Sprintf(SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD_COUNT,
		alertFilter.String(),
		recordFilter.String(),
		resultFilter.String(),
	)
}

func getSqlAndValueForSortedAlertEvent(req *request.AlertEventSearchRequest, cacheMinutes int) (string, []any, error) {
	alertFilter := NewQueryBuilder().
		Between("update_time", req.StartTime/1e6, req.EndTime/1e6).
		NotGreaterThan("end_time", req.EndTime/1e6)

	// TODO remove in v1.9.x
	{
		if len(req.Filter.Namespaces) > 0 {
			alertFilter.InStrings("tags['namespace']", req.Filter.Namespaces)
		}
		if len(req.Filter.Nodes) > 0 {
			alertFilter.InStrings("tags['node']", req.Filter.Nodes)
		}
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

	intervalMicro := int64(cacheMinutes) * int64(time.Minute) / 1e3
	endTime := req.EndTime/1e6 + int64(5*time.Minute)/1e9
	recordFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, endTime)

	resultFilter := NewQueryBuilder()
	// TODO remove in v1.9.x
	{
		if len(req.Filter.Validity) > 0 {
			resultFilter.InStrings("validity", req.Filter.Validity)
		}
		if len(req.Filter.Status) > 0 {
			resultFilter.InStrings("status", req.Filter.Status)
		}
	}

	err := applyFilter(req.Filters, resultFilter, alertFilter)
	if err != nil {
		return "", nil, err
	}

	sql := fmt.Sprintf(SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD,
		alertFilter.String(),
		recordFilter.String(),
		resultFilter.String(),
		resultOrder.String(),
	)

	values := make([]any, 0)
	values = append(values, alertFilter.values...)
	values = append(values, recordFilter.values...)
	values = append(values, resultFilter.values...)

	return sql, values, nil
}

func (ch *chRepo) GetAlertEventCounts(ctx core.Context, req *request.AlertEventSearchRequest, cacheMinutes int) (map[string]int64, error) {
	alertFilter := NewQueryBuilder().
		Between("update_time", req.StartTime/1e6, req.EndTime/1e6).
		NotGreaterThan("end_time", req.EndTime/1e6)

	var counts []_alertEventCount
	intervalMicro := int64(cacheMinutes) * int64(time.Minute) / 1e3
	endTime := req.EndTime/1e6 + int64(5*time.Minute)/1e9
	recordFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, endTime)
	countSql := fmt.Sprintf(SQL_GET_ALERTEVENT_COUNTS,
		alertFilter.String(),
		recordFilter.String(),
	)
	values := append(alertFilter.values, recordFilter.values...)
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &counts, countSql, values...)
	if err != nil {
		return nil, err
	}

	result := map[string]int64{
		"firing":         0,
		"resolved":       0,
		"valid":          0,
		"invalid":        0,
		"skipped":        0,
		"failed":         0,
		"unknown":        0,
		"firing-valid":   0,
		"firing-invalid": 0,
		"firing-other":   0,
	}

	for _, count := range counts {
		result[count.Status] += int64(count.Count)
		result[count.Validity] += int64(count.Count)

		if count.Status == alert.StatusFiring {
			if count.Validity == "valid" {
				result["firing-valid"] += int64(count.Count)
			} else if count.Validity == "invalid" {
				result["firing-invalid"] += int64(count.Count)
			} else {
				result["firing-other"] += int64(count.Count)
			}
		}
	}
	return result, nil
}

type _alertEventCount struct {
	Validity string `ch:"validity"`
	Status   string `ch:"status"`
	Count    uint64 `ch:"count"`
}
