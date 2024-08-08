package service

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// GetPolarisInfer implements Service.
func (s *service) GetPolarisInfer(req *request.GetPolarisInferRequest) (*response.GetPolarisInferResponse, error) {
	return s.polRepo.QueryPolarisInfer(
		req.StartTime, req.EndTime, prom.VecFromDuration(time.Duration(req.Step)*time.Microsecond),
		req.Service, req.Endpoint,
	)
}
