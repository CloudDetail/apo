package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServices(ctx core.Context, req *request.QueryServicesRequest) *response.QueryServicesResponse {
	services, err := s.promRepo.GetServiceList(ctx, req.StartTime, req.EndTime, []string{})
	if err != nil {
		return &response.QueryServicesResponse{
			Msg: "query services failed: " + err.Error(),
		}
	}
	results := make([]*response.QueryServiceResult, 0)
	for _, service := range services {
		results = append(results, &response.QueryServiceResult{
			Id:     service,
			Name:   service,
			Source: "apo",
		})
	}
	return &response.QueryServicesResponse{
		Results: results,
	}
}
