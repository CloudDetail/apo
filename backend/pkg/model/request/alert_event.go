// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

import "github.com/CloudDetail/apo/backend/pkg/model"

type AlertEventSearchRequest struct {
	StartTime int64 `json:"startTime" form:"startTime"`
	EndTime   int64 `json:"endTime" form:"endTime"`

	SortBy     string            `json:"sortBy" form:"sortBy"`
	Pagination *model.Pagination `json:"pagination"`

	Filters []AlertEventFilter `json:"filters,omitempty"`

	// Deprecated
	Filter AlertEventSearchFilter `json:"filter" form:"filter"`

	GroupID int64 `json:"groupId" form:"groupId"`

	SubGroupIDs []string `json:"-"`
}

type GetAlertDetailRequest struct {
	AlertID string `json:"alertId"`
	EventID string `json:"eventId"`

	StartTime  int64             `json:"startTime" form:"startTime"`
	EndTime    int64             `json:"endTime" form:"endTime"`
	Pagination *model.Pagination `json:"pagination"`

	LocateEvent bool `json:"locateEvent"`
}

// Deprecated: use AlertEventFilter instead
type AlertEventSearchFilter struct {
	Nodes      []string `json:"nodes" form:"nodes"`
	Namespaces []string `json:"namespaces" form:"namespaces"`

	// firing or resolved
	Status []string `json:"status" form:"status"`
	// valid or invalid or skipped or unknown
	Validity []string `json:"validity" form:"validity"`
}

type AlertEventClassifyRequest struct {
	AlertName  string `form:"alertName"`
	AlertGroup string `form:"alertGroup"`
}

type MarkAlertResolvedManuallyRequest struct {
	AlertID string `json:"alertId" form:"alertId"`
}

type SearchAlertEventFilterValuesRequest struct {
	Filters   []AlertEventFilter `json:"filters,omitempty"`
	SearchKey string             `json:"searchKey"`

	StartTime int64 `json:"startTime" form:"startTime"`
	EndTime   int64 `json:"endTime" form:"endTime"`
}

// AlertEventFilter
//
// Filtering based on the underlying fields and Tags of the AlertEvent itself
type AlertEventFilter struct {
	Name     string `json:"name"`
	Key      string `json:"key"`
	Wildcard bool   `json:"wildcard"`

	Options   []AlertEventFilterOption `json:"options,omitempty"`
	Selected  []string                 `json:"selected,omitempty"`
	MatchExpr string                   `json:"matchExpr,omitempty"`
}

type AlertEventFilterOption struct {
	Value   string `json:"value"`
	Display string `json:"display"`
}
