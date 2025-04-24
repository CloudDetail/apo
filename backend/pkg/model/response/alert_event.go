// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
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
