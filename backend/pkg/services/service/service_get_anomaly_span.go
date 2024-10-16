package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetAnomalySpan(req *request.GetAnomalySpanRequest) (response.GetAnomalySpanResponse, error) {
	if req.PageParam == nil {
		req.PageParam = &request.PageParam{
			CurrentPage: 1,
			PageSize:    999,
		}
	}
	resp := response.GetAnomalySpanResponse{}
	result, total, err := s.chRepo.GetAnomalyTrace(req)
	if err != nil {
		return resp, err
	}
	resp.List = result
	resp.Pagination = &model.Pagination{
		Total:       total,
		CurrentPage: req.CurrentPage,
		PageSize:    req.PageSize,
	}

	return resp, nil
}
