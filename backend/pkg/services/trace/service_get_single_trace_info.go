package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetSingleTraceID(req *request.GetSingleTraceInfoRequest) (*response.GetSingleTraceInfoResponse, error) {
	info, err := s.jaegerRepo.GetSingleTrace(req.TraceID)
	if err != nil {
		return nil, err
	}
	return &response.GetSingleTraceInfoResponse{
		TraceInfo: info,
	}, nil
}
