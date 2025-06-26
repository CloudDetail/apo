// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"encoding/json"
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/google/uuid"
)

const SQL_GET_LATEST_ALERT_EVENT_BY_ALERTID = `SELECT * FROM alert_event ae
%s ORDER BY received_time DESC LIMIT 1`

const SQL_GET_ALERT_DETAIL = `WITH targetEvent AS (
	SELECT *
    FROM alert_event ae
    %s
	LIMIT 1
),
latestEvent AS (
	SELECT *
	FROM alert_event ae
    %s
	ORDER BY received_time DESC limit 1
),
filterWorkflow AS(
    SELECT *,
      CASE
        WHEN output = 'false' THEN 2
        WHEN output = 'true' THEN 1
        ELSE 0
      END as importance
	FROM workflow_records
    %s
)
SELECT ae.id as id,
  ae.group as group,
  ae.name as name,
  ae.alert_id as alert_id,
  ae.create_time as create_time,
  ae.update_time as update_time,
  ae.end_time as end_time,
  ae.received_time as received_time,
  ae.detail as detail,
  ae.status as status,
  ae.severity as severity,
  ae.raw_tags as raw_tags,
  ae.tags as tags,
  ae.source as source,
  fw.workflow_run_id as workflow_run_id,
  fw.workflow_id as workflow_id,
  fw.workflow_name as workflow_name,
  fw.importance as importance,
  fw.output as output,
  fw.created_at as last_check_at,
  fw.alert_direction as alert_direction,
  fw.analyze_run_id as analyze_run_id,
  fw.analyze_err as analyze_err,
  le.status as last_status,
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
LEFT JOIN filterWorkflow fw ON toString(ae.id) = fw.input
LEFT JOIN latestEvent le ON ae.alert_id = le.alert_id
LIMIT 1`

const SQL_GET_RELEATED_ALERT_EVENT = `WITH filterEvent AS (
    SELECT *
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
),
filterNotify AS (
  SELECT *
  FROM alert_notify_record anr
  %s
)
SELECT ae.id as id,
  ae.group as group,
  ae.name as name,
  ae.alert_id as alert_id,
  ae.create_time as create_time,
  ae.update_time as update_time,
  ae.end_time as end_time,
  ae.received_time as received_time,
  ae.detail as detail,
  ae.status as status,
  ae.severity as severity,
  ae.tags as tags,
  ae.source as source,
  fw.workflow_run_id as workflow_run_id,
  fw.workflow_id as workflow_id,
  fw.workflow_name as workflow_name,
  fw.created_at as last_check_at,
  fw.importance as importance,
  fw.output as output,
  fw.alert_direction as alert_direction,
  fw.analyze_run_id as analyze_run_id,
  fw.analyze_err as analyze_err,
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
  END as validity,
  fn.success as notify_success,
  fn.failed as notify_failed,
  fn.created_at as notify_at
FROM filterEvent ae
LEFT JOIN filterWorkflow fw ON toString(ae.id) = fw.input
LEFT JOIN filterNotify fn ON toString(ae.id) = fn.event_id
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

func (ch *chRepo) GetAlertDetail(ctx core.Context, req *request.GetAlertDetailRequest, cacheMinutes int) (*alert.AEventWithWRecord, error) {
	alertFilter := NewQueryBuilder().
		Equals("alert_id", req.AlertID).
		Equals("toString(id)", req.EventID)

	lastEventFilter := NewQueryBuilder().
		Equals("alert_id", req.AlertID)

	recordFilter := NewQueryBuilder().
		Equals("ref", req.AlertID).
		Equals("input", req.EventID)

	sql := fmt.Sprintf(SQL_GET_ALERT_DETAIL,
		alertFilter.String(),
		lastEventFilter.String(),
		recordFilter.String(),
	)

	var values = alertFilter.values
	values = append(values, lastEventFilter.values...)
	values = append(values, recordFilter.values...)

	var result alert.AEventWithWRecord
	err := ch.GetContextDB(ctx).QueryRow(ctx.GetContext(), sql, values...).ScanStruct(&result)
	return &result, err
}

func (ch *chRepo) GetRelatedAlertEvents(ctx core.Context, req *request.GetAlertDetailRequest, cacheMinutes int) ([]alert.AEventWithWRecord, int64, error) {
	alertEventFilter := NewQueryBuilder().
		Between("update_time", req.StartTime/1e6, req.EndTime/1e6).
		NotGreaterThan("end_time", req.EndTime/1e6).
		Equals("alert_id", req.AlertID)

	countSql := fmt.Sprintf(SQL_GET_RELEATED_ALERT_EVENT_COUNT, alertEventFilter.String())

	var count uint64
	err := ch.GetContextDB(ctx).QueryRow(ctx.GetContext(), countSql, alertEventFilter.values...).Scan(&count)
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
		err := ch.GetContextDB(ctx).QueryRow(ctx.GetContext(), sql, values...).Scan(&index)
		if err != nil {
			return nil, 0, err
		}
		offSet = getOffset(int(index), req.Pagination.PageSize)
		req.Pagination.CurrentPage = int(index)/req.Pagination.PageSize + 1
	}

	intervalMicro := int64(cacheMinutes) * int64(time.Minute) / 1e3
	endTime := req.EndTime/1e6 + int64(5*time.Minute)/1e9
	recordFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, endTime).
		Equals("ref", req.AlertID)

	resultLimit := NewByLimitBuilder().
		OrderBy("received_time", false).
		Limit(req.Pagination.PageSize).Offset(offSet)

	notifyFilter := NewQueryBuilder().
		Between("created_at", (req.StartTime-intervalMicro)/1e6, endTime).
		Equals("alert_id", req.AlertID)

	sql := fmt.Sprintf(SQL_GET_RELEATED_ALERT_EVENT,
		alertEventFilter.String(),
		recordFilter.String(),
		notifyFilter.String(),
		resultLimit.String(),
	)

	values := alertEventFilter.values
	values = append(values, recordFilter.values...)
	values = append(values, notifyFilter.values...)

	var result = make([]alert.AEventWithWRecord, 0)
	err = ch.GetContextDB(ctx).Select(ctx.GetContext(), &result, sql, values...)
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

func (ch *chRepo) GetLatestAlertEventByAlertID(ctx core.Context, alertID string) (*alert.AlertEvent, error) {
	alertEventFilter := NewQueryBuilder().
		Equals("alert_id", alertID).
		GreaterThan("received_time", time.Now().Add(-7*24*time.Hour).Unix())
	sql := fmt.Sprintf(SQL_GET_LATEST_ALERT_EVENT_BY_ALERTID, alertEventFilter.String())
	var result = alert.AlertEvent{}
	err := ch.GetContextDB(ctx).QueryRow(ctx.GetContext(), sql, alertEventFilter.values...).ScanStruct(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

func (ch *chRepo) ManualResolveLatestAlertEventByAlertID(ctx core.Context, alertID string) error {
	alertEventFilter := NewQueryBuilder().
		Equals("alert_id", alertID).
		GreaterThan("received_time", time.Now().Add(-7*24*time.Hour).Unix())
	sql := fmt.Sprintf(SQL_GET_LATEST_ALERT_EVENT_BY_ALERTID, alertEventFilter.String())
	var result = alert.AlertEvent{}
	err := ch.GetContextDB(ctx).QueryRow(ctx.GetContext(), sql, alertEventFilter.values...).ScanStruct(&result)
	if err != nil {
		return err
	}

	if result.Status == model.StatusResolved.ToString() {
		return nil
	}

	now := time.Now()
	result.Status = model.StatusResolved.ToString()
	result.EndTime = now
	result.ReceivedTime = now
	result.ID = uuid.New()

	detail := map[string]string{
		"description": fmt.Sprintf("alert has been closed manally, LABELS: %+v", result.Tags),
	}

	detailsStr, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	result.Detail = string(detailsStr)

	return ch.InsertAlertEvent(ctx, []alert.AlertEvent{result}, alert.SourceFrom{
		SourceInfo: alert.SourceInfo{SourceName: result.Source},
	})
}
