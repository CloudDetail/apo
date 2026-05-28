// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"encoding/json"
	"fmt"
	"sort"
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
)

func querytagsCondition(tags map[string]string) string {
	if len(tags) == 0 {
		return "AND (1='1')"
	}

	conditions := make([]string, 0, len(tags))
	for k, v := range tags {
		conditions = append(conditions, fmt.Sprintf(
			"tags['%s'] = '%s'",
			escapeClickHouseStringLiteral(k),
			escapeClickHouseStringLiteral(v),
		))
	}
	sort.Strings(conditions)
	return "AND " + strings.Join(conditions, " AND ")
}

func escapeClickHouseStringLiteral(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(escaped, "'", "''")
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

func (ch *chRepo) ListAlertEventReports(ctx core.Context, startTime time.Time, endTime time.Time, limit int) ([]model.AlertEventReport, error) {
	if limit <= 0 {
		limit = 20
	}
	builder := NewQueryBuilder().
		Between("overview_timestamp", startTime.Unix(), endTime.Unix())

	byLimit := NewByLimitBuilder().
		OrderBy("overview_timestamp", false).
		Limit(limit)

	sql := fmt.Sprintf(SQL_GET_ALERT_EVENT_REPORT, builder.String(), "", byLimit.String())

	var reports []model.AlertEventReport
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &reports, sql, startTime, endTime)
	return reports, err
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
		model.ReportOverviewTagsForStorage(req.Overview.Tags, req.Overview.Headline, req.ReportView),
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
