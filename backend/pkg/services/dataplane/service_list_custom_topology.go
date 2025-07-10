package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) ListCustomTopology(ctx core.Context, req *request.ListCustomTopologyRequest) (*response.ListCustomTopologyResponse, error) {
	topologies, err := s.dbRepo.ListCustomServiceTopology(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*database.CustomServiceTopology, 0)
	for _, topology := range topologies {
		if topology.StartTime > 0 && topology.ExpireTime > 0 &&
			(req.StartTime > topology.ExpireTime || req.EndTime < topology.StartTime) {
			continue
		}
		result = append(result, &topology)
	}
	return &response.ListCustomTopologyResponse{
		Topologies: result,
	}, nil
}