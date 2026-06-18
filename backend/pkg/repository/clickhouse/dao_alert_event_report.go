// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	SQL_GET_ALERT_EVENT_REPORT = `
	SELECT report_type,overview_timestamp,overview_detail,overview_reason,overview_tags,tags,topology,rootcause,rootcause,rootcause_other,suggest,data
	FROM alert_report %s %s %s`

	SQL_GET_ALERT_EVENT_REPORT_BY_ANALYZE_RUN_ID = `
	SELECT
		ar.report_type AS report_type,
		ar.overview_timestamp AS overview_timestamp,
		ar.overview_detail AS overview_detail,
		ar.overview_reason AS overview_reason,
		ar.overview_tags AS overview_tags,
		ar.tags AS tags,
		ar.topology AS topology,
		ar.rootcause AS rootcause,
		ar.rootcause_other AS rootcause_other,
		ar.suggest AS suggest,
		ar.data AS data
	FROM workflow_records wr
	INNER JOIN alert_report ar ON ar.tags['alertEventId'] = wr.input
	WHERE wr.analyze_run_id = ?
	  AND ar.overview_timestamp BETWEEN ? AND ?
	ORDER BY ar.overview_timestamp DESC
	LIMIT 1`

	SQL_GET_ALERT_EVENT_REPORTS_BY_ALERT_ID = `
	SELECT
		ar.report_type AS report_type,
		ar.overview_timestamp AS overview_timestamp,
		ar.overview_detail AS overview_detail,
		ar.overview_reason AS overview_reason,
		ar.overview_tags AS overview_tags,
		ar.tags AS tags,
		ar.topology AS topology,
		ar.rootcause AS rootcause,
		ar.rootcause_other AS rootcause_other,
		ar.suggest AS suggest,
		ar.data AS data
	FROM (
		SELECT DISTINCT input
		FROM workflow_records
		WHERE ref = ?
		  AND created_at BETWEEN ? AND ?
		  AND input != ''
	) wr
	INNER JOIN alert_report ar ON ar.tags['alertEventId'] = wr.input
	WHERE ar.overview_timestamp BETWEEN ? AND ?
	ORDER BY ar.overview_timestamp DESC`
)

func querytagsCondition(tags map[string]string) string {
	var res string
	for k, v := range tags {
		res += fmt.Sprintf("AND tags['%s'] = '%s'", k, v)
	}
	if res == "" {
		res = "AND (1='1')"
	}
	return res
}

func (ch *chRepo) GetAlertEventReport(ctx core.Context, startTime time.Time, endTime time.Time, tags map[string]string) (*model.AlertEventReport, error) {
	builder := NewQueryBuilder().
		Between("overview_timestamp", startTime.Unix(), endTime.Unix())

	sqltags := querytagsCondition(tags)
	byLimit := NewByLimitBuilder().Limit(1)

	sql := fmt.Sprintf(SQL_GET_ALERT_EVENT_REPORT, builder.String(), sqltags, byLimit.String())

	var event []model.AlertEventReport
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &event, sql, startTime, endTime)
	if err != nil {
		return nil, err
	} else if len(event) == 0 {
		return nil, fmt.Errorf("no event found with tags %s", sqltags)
	}
	return &event[0], err
}

func (ch *chRepo) GetAlertEventReportByAnalyzeRunID(ctx core.Context, startTime time.Time, endTime time.Time, analyzeRunID string) (*model.AlertEventReport, error) {
	analyzeRunID = strings.TrimSpace(analyzeRunID)
	if analyzeRunID == "" {
		return nil, fmt.Errorf("empty analyze_run_id")
	}
	var reports []model.AlertEventReport
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &reports, SQL_GET_ALERT_EVENT_REPORT_BY_ANALYZE_RUN_ID, analyzeRunID, startTime, endTime)
	if err != nil {
		return nil, err
	} else if len(reports) == 0 {
		return nil, fmt.Errorf("no alert report found with analyze_run_id = '%s'", analyzeRunID)
	}
	return &reports[0], nil
}

func (ch *chRepo) GetAlertEventReportsByAlertID(ctx core.Context, startTime time.Time, endTime time.Time, alertID string) ([]model.AlertEventReport, error) {
	alertID = strings.TrimSpace(alertID)
	if alertID == "" {
		return nil, fmt.Errorf("empty alert_id")
	}
	var reports []model.AlertEventReport
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &reports, SQL_GET_ALERT_EVENT_REPORTS_BY_ALERT_ID, alertID, startTime, endTime, startTime, endTime)
	if err != nil {
		return nil, err
	} else if len(reports) == 0 {
		return nil, fmt.Errorf("no alert report found with alert_id = '%s'", alertID)
	}
	return reports, nil
}

func convertJSONS2Strings(datas []map[string]any) []string {
	result := make([]string, 0, len(datas))
	for _, data := range datas {
		result = append(result, convertJSON2String(data))
	}
	return result
}

func convertJSON2String(data map[string]any) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

func (ch *chRepo) AddAlertEventReport(ctx core.Context, req *request.AddAlertEventReportRequest) error {

	batch, err := ch.GetContextDB(ctx).PrepareBatch(ctx.GetContext(), `
			INSERT INTO alert_report (
    	report_type,
    	overview_timestamp,
    	overview_detail,
    	overview_reason,
    	overview_tags,
    	tags,
    	topology,
    	rootcause,
    	rootcause_other,
    	suggest,
    	data
		) VALUES
	`)
	if err != nil {
		return err
	}
	if err := batch.Append(
		req.Type,
		time.UnixMicro(req.Overview.TimeStamp),
		req.Overview.Detail,
		req.Overview.Reason,
		req.Overview.Tags,
		req.Tags,
		convertJSON2String(req.Topology),
		req.RootCause.Cause,
		convertJSONS2Strings(req.RootCause.Other),
		convertJSONS2Strings(req.Suggest),
		convertJSONS2Strings(req.Data),
	); err != nil {
		return err
	}
	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}
