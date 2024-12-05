package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetServiceList(req *request.GetServiceListRequest) ([]string, error) {
	return s.promRepo.GetServiceList(req.StartTime, req.EndTime, req.Namespace)
}
