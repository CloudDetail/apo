// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"encoding/json"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

type AlertEventSearchResponse struct {
	EventList  []alert.AEventWithWRecord `json:"events"`
	Pagination *model.Pagination         `json:"pagination"`

	AlertEventAnalyzeWorkflowID string           `json:"alertEventAnalyzeWorkflowId"`
	AlertCheckID                string           `json:"alertCheckId"`
	Counts                      map[string]int64 `json:"counts"`
}

type GetAlertDetailResponse struct {
	CurrentEvent *alert.AEventWithWRecord `json:"currentEvent"`

	EventList  []alert.AEventWithWRecord `json:"events"`
	Pagination *model.Pagination         `json:"pagination"`

	AlertEventAnalyzeWorkflowID string `json:"alertEventAnalyzeWorkflowId"`
	AlertCheckID                string `json:"alertCheckId"`

	LocateIdx int `json:"locateIndex"`
}

type AlertEventClassifyResponse struct {
	WorkflowId string `json:"workflowId"`
}

type AlertEventFiltersResponse struct {
	Filters []request.AlertEventFilter `json:"filters"`
}

type AlertEventFilterLabelKeysResponse struct {
	Labels []string `json:"labels"`
}

type GetAlertEventReportResponse struct {
	Type string `json:"reportType"`
	Data any    `json:"reportData"`
}

type GetAlertEventReportViewResponse struct {
	Snapshot GetAlertEventReportResponse `json:"snapshot"`
	View     ReportView                  `json:"view"`
}

type ReportView struct {
	ReportID       string                `json:"reportId,omitempty"`
	Headline       string                `json:"headline,omitempty"`
	AnalysisWindow *ReportAnalysisWindow `json:"analysisWindow,omitempty"`
	RuntimeStatus  *ReportRuntimeStatus  `json:"runtimeStatus,omitempty"`
	BusinessImpact *ReportBusinessImpact `json:"businessImpact,omitempty"`
	MetricDeltas   []ReportMetricDelta   `json:"metricDeltas,omitempty"`
	Topology       *map[string]any       `json:"topology,omitempty"`
	FaultLocation  *ReportFaultLocation  `json:"faultLocation,omitempty"`
	Timeline       []ReportTimelineItem  `json:"timeline,omitempty"`
	DataSources    ReportViewDataSources `json:"dataSources"`
}

type ReportAnalysisWindow struct {
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
}

type ReportRuntimeStatus struct {
	State     string `json:"state"`
	Label     string `json:"label"`
	Summary   string `json:"summary"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
}

type ReportBusinessImpact struct {
	Summary          string   `json:"summary"`
	Endpoints        []string `json:"endpoints"`
	EntryServices    []string `json:"entryServices"`
	AffectedServices []string `json:"affectedServices"`
}

type ReportMetricDelta struct {
	Name        string `json:"name"`
	Before      any    `json:"before,omitempty"`
	Current     any    `json:"current,omitempty"`
	Unit        string `json:"unit,omitempty"`
	Description string `json:"description,omitempty"`
}

type ReportFaultLocation struct {
	Label       string   `json:"label"`
	Service     string   `json:"service"`
	Description string   `json:"description"`
	Path        []string `json:"path"`
	Basis       string   `json:"basis"`
}

type ReportTimelineItem struct {
	Time   string `json:"time,omitempty"`
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
}

type ReportViewDataSources struct {
	Snapshot       string `json:"snapshot"`
	RuntimeStatus  string `json:"runtimeStatus,omitempty"`
	BusinessImpact string `json:"businessImpact,omitempty"`
	MetricDeltas   string `json:"metricDeltas,omitempty"`
	Topology       string `json:"topology,omitempty"`
	Timeline       string `json:"timeline,omitempty"`
}

func convertJSONStringsToJSON(jsonStrings []string) []map[string]any {
	result := make([]map[string]any, 0, len(jsonStrings))

	for _, str := range jsonStrings {
		var v map[string]any
		err := json.Unmarshal([]byte(str), &v)
		if err != nil {
			continue
		}
		result = append(result, v)
	}

	return result
}

func deserializeTopology(jsonStr string) *map[string]any {
	if jsonStr == "" || jsonStr == "{}" {
		return nil
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil
	}

	return &result
}

func ToAlertReportResponse(report model.AlertEventReport) GetAlertEventReportResponse {
	var data ReportData

	data.OverView.Reason = report.OverviewDetail
	data.AnalyzeProcess = report.OverviewReason
	data.OverView.Tags = model.ReportOverviewPublicTags(report.OverviewTags)
	data.OverView.TimeStamp = report.OverviewTimestamp.UnixMicro()
	data.ReportView = model.ReportViewFromTags(report.OverviewTags)

	data.Topology = deserializeTopology(report.Topology)

	data.RootCause.Cause = report.RootCause
	data.RootCause.Other = convertJSONStringsToJSON(report.RootCauseOther)

	data.Suggest = convertJSONStringsToJSON(report.Suggest)
	data.Data = convertJSONStringsToJSON(report.Data)

	return GetAlertEventReportResponse{
		Data: data,
		Type: report.ReportType,
	}
}

func ToAlertReportViewResponse(report model.AlertEventReport, req *request.GetAlertEventReportRequest) GetAlertEventReportViewResponse {
	snapshot := toAlertReportSnapshotResponse(report)
	view := ReportView{
		DataSources: ReportViewDataSources{
			Snapshot: "alert_report",
		},
	}

	return GetAlertEventReportViewResponse{
		Snapshot: snapshot,
		View:     view,
	}
}

func toAlertReportSnapshotResponse(report model.AlertEventReport) GetAlertEventReportResponse {
	var data ReportData

	data.OverView.Detail = report.OverviewDetail
	data.OverView.Reason = report.OverviewReason
	data.OverView.Tags = model.ReportOverviewPublicTags(report.OverviewTags)
	data.OverView.TimeStamp = report.OverviewTimestamp.UnixMicro()
	data.ReportView = model.ReportViewFromTags(report.OverviewTags)

	data.Topology = deserializeTopology(report.Topology)

	data.RootCause.Cause = report.RootCause
	data.RootCause.Other = convertJSONStringsToJSON(report.RootCauseOther)

	data.Suggest = convertJSONStringsToJSON(report.Suggest)
	data.AnalyzeProcess = report.OverviewReason
	data.Data = convertJSONStringsToJSON(report.Data)

	return GetAlertEventReportResponse{
		Data: data,
		Type: report.ReportType,
	}
}

type ReportData struct {
	OverView       model.ReportOverView  `json:"overview"`
	Topology       *map[string]any       `json:"topology"`
	RootCause      model.ReportRootCause `json:"rootCauseAnalysis"`
	Suggest        []map[string]any      `json:"suggest"`
	AnalyzeProcess string                `json:"analyzeProcess"`
	Data           []map[string]any      `json:"evidence"`
	ReportView     map[string]any        `json:"reportView,omitempty"`
}
