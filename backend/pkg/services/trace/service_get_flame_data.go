package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetFlameGraphData(req *request.GetFlameDataRequest) (response.GetFlameDataResponse, error) {
	flameData, err := s.chRepo.GetFlameGraphData(req.StartTime, req.EndTime, req.PID, req.TID, req.SampleType, req.SpanID, req.TraceID)
	if err != nil {
		return response.GetFlameDataResponse{}, err
	}
	return (response.GetFlameDataResponse)((*flameData)[0]), nil
}
