// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

import "github.com/CloudDetail/apo/backend/pkg/model/customdashboard"

type DashboardListRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
}

type DashboardIDRequest struct {
	ID string `form:"id" json:"id" binding:"required"`
}

type CreateDashboardRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TemplateID  string `json:"templateId"`
	Blank       bool   `json:"blank"`
}

type UpdateDashboardRequest struct {
	customdashboard.Dashboard
}

type DashboardVariableRequest struct {
	Type          string              `form:"type" json:"type" binding:"required"`
	Label         string              `form:"label" json:"label"`
	VariableName  string              `form:"variableName" json:"variableName"`
	Query         string              `form:"query" json:"query"`
	GroupID       *int64              `form:"groupId" json:"groupId"`
	Variables     map[string][]string `form:"variables" json:"variables"`
	VariablesJSON string              `form:"variablesJson" json:"variablesJson"`
	StartTime     int64               `form:"startTime" json:"startTime"`
	EndTime       int64               `form:"endTime" json:"endTime"`
}

type DashboardQueryRequest struct {
	DashboardID string               `json:"dashboardId"`
	WidgetID    string               `json:"widgetId"`
	Mode        string               `json:"mode"`
	PresetID    string               `json:"presetId"`
	Expr        string               `json:"expr"`
	Queries     []DashboardQueryItem `json:"queries"`
	Params      map[string]string    `json:"params"`
	Variables   map[string][]string  `json:"variables"`
	GroupID     *int64               `json:"groupId"`
	ChartType   string               `json:"chartType"`
	SeriesLimit int                  `json:"seriesLimit"`
	StartTime   int64                `json:"startTime"`
	EndTime     int64                `json:"endTime"`
	Step        int64                `json:"step"`
	Instant     bool                 `json:"instant"`
}

type DashboardQueryItem struct {
	ID         string `json:"id"`
	Expression string `json:"expression"`
	Legend     string `json:"legend"`
	Hidden     bool   `json:"hidden,omitempty"`
}

type DashboardValidateQueryRequest struct {
	Expr string `json:"expr" binding:"required"`
}
