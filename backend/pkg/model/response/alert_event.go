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
	if jsonStr == "" {
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

	data.OverView.Detail = report.OverviewDetail
	data.OverView.Reason = report.OverviewReason
	data.OverView.Tags = report.OverviewTags
	data.OverView.TimeStamp = report.OverviewTimestamp.UnixMicro()

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

type ReportData struct {
	OverView  model.ReportOverView  `json:"overview"`
	Topology  *map[string]any       `json:"topology"`
	RootCause model.ReportRootCause `json:"rootCauseAnalysis"`
	Suggest   []map[string]any      `json:"suggest"`
	Data      []map[string]any      `json:"evidence"`
}
