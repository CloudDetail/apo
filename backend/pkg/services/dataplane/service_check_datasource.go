package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (s *service) CheckDataSource(ctx core.Context, req *model.CheckDataSourceRequest) *model.CheckDataSourceResponse {
	resp, err := s.dataplaneRepo.CheckDataSource(req)
	if err != nil {
		return &model.CheckDataSourceResponse{
			Success: false,
			Message: err.Error(),
		}
	}
	return resp
}
