// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

import "github.com/CloudDetail/apo/backend/pkg/model"

type AlertEventSearchRequest struct {
	StartTime int64 `json:"startTime" form:"startTime"`
	EndTime   int64 `json:"endTime" form:"endTime"`

	SortBy     string            `json:"sortBy" form:"sortBy"`
	Pagination *model.Pagination `json:"pagination"`

	Filter AlertEventSearchFilter `json:"filter" form:"filter"`
}

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
