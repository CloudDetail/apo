package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetOnOffCPU(req *request.GetOnOffCPURequest) (*response.GetOnOffCPUResponse, error) {
	result, err := s.chRepo.GetOnOffCPU(req.PID, req.NodeName, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	resp := &response.GetOnOffCPUResponse{
		ProfilingEvent: result,
	}
	return resp, nil
}
