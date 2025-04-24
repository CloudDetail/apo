package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const SQL_GET_ALERT_DETAIL = `WITH targetEvent AS (
	SELECT *, %s as rounded_time
    FROM alert_event ae
    %s
	LIMIT 1
),
filterWorkflow AS(
    SELECT *,
      CASE
        WHEN output = 'false' THEN 2
        WHEN output = 'true' THEN 1
        ELSE 0
      END as importance
	FROM workflow_records
    %s AND rounded_time = (SELECT rounded_time FROM targetEvent)
)
SELECT ae.id,
  ae.group,
  ae.name,
  ae.alert_id,
  ae.create_time,
  ae.update_time,
  ae.end_time,
  ae.received_time,
  ae.detail,
  ae.status,
  ae.raw_tags,
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
FROM targetEvent ae
LEFT JOIN filterWorkflow fw ON ae.rounded_time = fw.rounded_time
LIMIT 1`

const SQL_GET_RELEATED_ALERT_EVENT = `WITH filterEvent AS (
    SELECT *, %s as rounded_time
    FROM alert_event ae
    %s
),
filterWorkflow AS (
  SELECT *,
    CASE
      WHEN output = 'false' THEN 2
      WHEN output = 'true' THEN 1
      ELSE 0
    END as importance
  FROM workflow_records fw
  %s
)
SELECT ae.id,
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
  fw.created_at,
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
FROM filterEvent ae
LEFT JOIN filterWorkflow fw ON ae.rounded_time = fw.rounded_time
%s`

const SQL_GET_RELEATED_ALERT_EVENT_COUNT = `SELECT count(1) AS count
FROM alert_event ae
%s`

const SQL_GET_RELEATED_ALERT_EVENT_LOCATE = `WITH target_event AS (
    SELECT alert_id,received_time FROM alert_event ae
    %s
    LIMIT 1
)
SELECT count(1) as index
FROM alert_event ae
%s AND received_time > (SELECT received_time FROM target_event)`

func (ch *chRepo) GetAlertDetail(req *request.GetAlertDetailRequest, cacheMinutes int) (*alert.AEventWithWRecord, error) {
	recordFilter := NewQueryBuilder().
		Equals("ref", req.AlertID)

	alertFilter := NewQueryBuilder().
		Equals("alert_id", req.AlertID).
		Equals("toString(id)", req.EventID)

	sql := fmt.Sprintf(SQL_GET_ALERT_DETAIL,
		getEventRoundedTime(cacheMinutes),
		alertFilter.String(),
		recordFilter.String(),
	)

	var values = alertFilter.values
	values = append(values, recordFilter.values...)

	var result alert.AEventWithWRecord
	err := ch.conn.QueryRow(context.Background(), sql, values...).ScanStruct(&result)
	return &result, err
}

func (ch *chRepo) GetRelatedAlertEvents(req *request.GetAlertDetailRequest, cacheMinutes int) ([]alert.AEventWithWRecord, int64, error) {
	alertEventFilter := NewQueryBuilder().
		Between("update_time", req.StartTime/1e6, req.EndTime/1e6).
		NotGreaterThan("end_time", req.EndTime/1e6).
		Equals("alert_id", req.AlertID)

	countSql := fmt.Sprintf(SQL_GET_RELEATED_ALERT_EVENT_COUNT, alertEventFilter.String())

	var count uint64
	err := ch.conn.QueryRow(context.Background(), countSql, alertEventFilter.values...).Scan(&count)
	if err != nil || count == 0 {
		return nil, int64(count), err
	}

	var offSet = (req.Pagination.CurrentPage - 1) * req.Pagination.PageSize
	if req.LocateEvent {
		targetQuery := NewQueryBuilder().
			Equals("toString(id)", req.EventID).
			Equals("alert_id", req.AlertID)

		sql := fmt.Sprintf(SQL_GET_RELEATED_ALERT_EVENT_LOCATE, targetQuery.String(), alertEventFilter.String())
		var values = targetQuery.values
		values = append(values, alertEventFilter.values...)

		var index uint64
		err := ch.conn.QueryRow(context.Background(), sql, values...).Scan(&index)
		if err != nil {
			// TODO do something
		}
		offSet = getOffset(int(index), req.Pagination.PageSize)
		req.Pagination.CurrentPage = int(index)/req.Pagination.PageSize + 1
	}

	var result = make([]alert.AEventWithWRecord, 0)
	intervalMicro := int64(5*time.Minute) / 1e3
	recordFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, (req.EndTime+intervalMicro)/1e6).
		Equals("ref", req.AlertID)

	resultLimit := NewByLimitBuilder().
		OrderBy("received_time", false).
		Limit(req.Pagination.PageSize).Offset(offSet)

	sql := fmt.Sprintf(SQL_GET_RELEATED_ALERT_EVENT,
		getEventRoundedTime(cacheMinutes),
		alertEventFilter.String(),
		recordFilter.String(),
		resultLimit.String(),
	)

	values := alertEventFilter.values
	values = append(values, recordFilter.values...)
	err = ch.conn.Select(context.Background(), &result, sql, values...)
	return result, int64(count), err
}

func getOffset(rowIndex, pageSize int) (offset int) {
	if pageSize <= 0 {
		return 0
	}
	pageNumber := rowIndex / pageSize
	offset = pageNumber * pageSize
	return offset
}
