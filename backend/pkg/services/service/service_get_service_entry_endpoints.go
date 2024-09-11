package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceEntryEndpoints(req *request.GetServiceEntryEndpointsRequest) (*response.GetServiceEntryEndpointsResponse, error) {
	entries, err := s.chRepo.ListEntryEndpoints(req)
	if err != nil {
		return nil, err
	}

	services := make([]string, len(entries))
	endpoints := make([]string, len(entries))
	for i, entry := range entries {
		services[i] = entry.Service
		endpoints[i] = entry.Endpoint
	}

	return &response.GetServiceEntryEndpointsResponse{
		Services:  services,
		EndPoints: endpoints,
	}, nil
}
