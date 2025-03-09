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

func sortbyParam(sortBy string) ([]string, []bool) {
	if len(sortBy) == 0 {
		return []string{"received_time"}, []bool{true}
	}
	sortBys := strings.Split(sortBy, ",")
	if len(sortBys) == 0 {
		return []string{"received_time"}, []bool{true}
	}

	var fields []string
	var ascs []bool
	for _, option := range sortBys {
		parts := strings.Split(option, " ")
		if len(parts) == 1 {
			fields = append(fields, parts[0])
			ascs = append(ascs, true)
		} else if len(parts) == 2 {
			fields = append(fields, parts[0])
			ascs = append(ascs, parts[1] == "asc")
		} else {
			// TODO 无效的过滤
			continue
		}
	}

	return fields, ascs
}

func (ch *chRepo) GetAlertEventWithWorkflowRecord(req *request.AlertEventSearchRequest) ([]alert.AEventWithWRecord, int64, error) {
	alertFilter := NewQueryBuilder().
		Between("received_time", req.StartTime/1e6, req.EndTime/1e6)

	var count uint64
	countSql := fmt.Sprintf(SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD_COUNT, alertFilter.String())
	err := ch.conn.QueryRow(context.Background(), countSql, alertFilter.values...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	alertOrder := NewByLimitBuilder()
	fields, ascs := sortbyParam(req.SortBy)
	for idx, field := range fields {
		alertOrder.OrderBy(field, ascs[idx])
	}
	if req.Pagination != nil {
		alertOrder.Offset((req.Pagination.CurrentPage - 1) * req.Pagination.PageSize).
			Limit(req.Pagination.PageSize)
	}

	intervalMicro := int64(5*time.Minute) / 1e3
	recordFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, (req.EndTime+intervalMicro)/1e6)

	finalOrder := NewByLimitBuilder()
	for idx, field := range fields {
		finalOrder.OrderBy(field, ascs[idx])
	}

	sql := fmt.Sprintf(SQL_GET_ALERTEVENT_WITH_WORKFLOW_RECORD,
		alertFilter.String(), alertOrder.String(),
		recordFilter.String(),
		finalOrder.String(),
	)

	result := make([]alert.AEventWithWRecord, 0)
	values := make([]any, 0)
	values = append(values, alertFilter.values...)
	values = append(values, recordFilter.values...)
	err = ch.conn.Select(context.Background(), &result, sql, values...)
	return result, int64(count), err
}
