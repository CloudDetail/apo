package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) AlertDetail(req *request.GetAlertDetailRequest) (*response.GetAlertDetailResponse, error) {
	eventDetail, err := s.chRepo.GetAlertDetail(req, s.alertWorkflow.AlertCheck.CacheMinutes)
	if err != nil {
		// TODO
	}

	releatedEvents, total, err := s.chRepo.GetRelatedAlertEvents(req, s.alertWorkflow.AlertCheck.CacheMinutes)
	if err != nil {
		// TODO
	}

	req.Pagination.Total = total
	var localIndex int
	for idx, event := range releatedEvents {
		if event.ID.String() == req.EventID {
			localIndex = idx
		}
	}

	return &response.GetAlertDetailResponse{
		CurrentEvent:                eventDetail,
		EventList:                   releatedEvents,
		Pagination:                  req.Pagination,
		AlertEventAnalyzeWorkflowID: s.alertWorkflow.AnalyzeFlowId,
		AlertCheckID:                s.alertWorkflow.AlertCheck.FlowId,
		LocateIdx:                   localIndex,
	}, nil
}
