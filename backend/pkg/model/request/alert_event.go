package request

import "github.com/CloudDetail/apo/backend/pkg/model"

type AlertEventSearchRequest struct {
	StartTime int64 `json:"startTime" form:"startTime"`
	EndTime   int64 `json:"endTime" form:"endTime"`

	SortBy     string            `json:"sortBy" form:"sortBy"`
	Pagination *model.Pagination `json:"pagination"`

	WorkflowParams map[string]string `json:"workflowParams"`
}
