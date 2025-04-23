package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const SQL_GET_ALERT_DETAIL = `WITH filterWorkflow AS(
	SELECT * FROM workflow_record
	%s
)
SELECT * FROM alert_event ae
LEFT JOIN filterWorkflow fw ON ae.alert_id = fw.ref AND %s = fw.rounded_time
%s`

const SQL_GET_RELEATED_ALERT_EVENT = `WITH filterEvent AS (
	SELECT *, %s as rount_time
	FROM alert_event ae
	%s
),
filterWorkflow AS (
	SELECT * FROM workflow_records fw
	%s
)
SELECT *
FROM filterEvent ae LEFT JOIN filterWorkflow fw ON ae.round_time = fw.rount_time
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
%s AND receiver_time < (SELECT received_time FROM target_event)`

func (ch *chRepo) GetAlertDetail(req *request.GetAlertDetailRequest, cacheMinutes int) (*alert.AEventWithWRecord, error) {
	intervalMicro := int64(5*time.Minute) / 1e3
	recordFilter := NewQueryBuilder().
		Equals("ref", req.AlertID).
		Between("created_at", (req.StartTime-intervalMicro)/1e6, (req.EndTime+intervalMicro)/1e6)

	alertFilter := NewQueryBuilder().
		Equals("alert_id", req.AlertID).
		Equals("toString(id)", req.EventID)

	sql := fmt.Sprintf(SQL_GET_ALERT_DETAIL,
		recordFilter.String(),
		getEventRoundedTime(cacheMinutes),
		alertFilter.String(),
	)

	var values = recordFilter.values
	values = append(values, alertFilter.values...)

	var result alert.AEventWithWRecord
	err := ch.conn.QueryRow(context.Background(), sql, values...).ScanStruct(&result)
	return &result, err
}

func (ch *chRepo) GetRelatedAlertEvents(req *request.GetAlertDetailRequest, cacheMinutes int) ([]alert.AEventWithWRecord, int64, error) {
	alertEventFilter := NewQueryBuilder().
		Between("update_time", req.StartTime/1e6, req.EndTime/1e6).
		NotGreaterThan("end_time", req.EndTime/1e6).
		Equals("alert_id", req.AlertID)

	var count uint64
	err := ch.conn.QueryRow(context.Background(), alertEventFilter.String(), alertEventFilter.values...).Scan(&count)
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
		err := ch.conn.QueryRow(context.Background(), sql, values...).Scan(index)
		if err != nil {
			// TODO do something
		}
		offSet = getOffset(int(index), req.Pagination.PageSize)
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
		alertEventFilter.String(),
		getEventRoundedTime(cacheMinutes),
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
