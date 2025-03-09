package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type AlertEventSearchResponse struct {
	EventList  []alert.AEventWithWRecord `json:"events"`
	Pagination *model.Pagination         `json:"pagination"`

	AlertEventAnalyzeWorkflowID string `json:"alertEventAnalyzeWorkflowId"`
}
