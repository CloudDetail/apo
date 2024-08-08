package mock

import (
	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/internal/model/response"
)

func (s *service) PageList(req *request.ListRequest) (resp *response.ListResponse, err error) {
	list, count, err := s.dbRepo.ListMocksByCondition(req)
	if err != nil {
		return nil, err
	}

	return &response.ListResponse{
		List: list,
		Pagination: &response.Pagination{
			Total:        count,
			CurrentPage:  req.PageNum,
			PerPageCount: req.PageSize,
		},
	}, nil
}
