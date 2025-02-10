package alertanalyze

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetDefectDetectExecList(req *request.GetDefectDetectExecListRequest) (*response.GetDefectDetectExecListResponse, error) {
	list, totalCount, err := s.chRepo.GetDetectExecList(req)
	if err != nil {
		return nil, err
	}
	resp := &response.GetDefectDetectExecListResponse{
		Pagination: model.Pagination{
			CurrentPage: req.CurrentPage,
			Total:       totalCount,
			PageSize:    req.PageSize,
		},
		MutatedRule: list,
	}

	return resp, nil
}
