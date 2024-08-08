package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetTracePageList(req *request.GetTracePageListRequest) (*response.GetTracePageListResponse, error) {
	list, total, err := s.chRepo.GetTracePageList(req)
	if err != nil {
		return nil, err
	}
	return &response.GetTracePageListResponse{
		Pagination: &model.Pagination{
			Total:       total,
			CurrentPage: req.PageNum,
			PageSize:    req.PageSize,
		},
		List: list,
	}, nil
}
