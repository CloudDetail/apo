package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) CreateCustomTopology(ctx core.Context, req *request.CreateCustomTopologyRequest) error {
	return s.dbRepo.CreateCustomServiceTopology(ctx, &database.CustomServiceTopology{
		ClusterId: req.ClusterId,
		LeftNode: req.LeftNode,
		LeftType: req.LeftType,
		RightNode: req.RightNode,
		RightType: req.RightType,
	})
}