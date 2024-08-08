package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetFaultLogContent implements Service.
func (s *service) GetFaultLogContent(req *request.GetFaultLogContentRequest) (*response.GetFaultLogContentResponse, error) {
	logContest, sources, err := s.chRepo.QueryApplicationLogs(req)
	if err != nil {
		return nil, err
	}
	return &response.GetFaultLogContentResponse{
		Sources:     sources,
		LogContents: logContest,
	}, nil
}
