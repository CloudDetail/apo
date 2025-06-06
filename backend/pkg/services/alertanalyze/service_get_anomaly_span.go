package alertanalyze

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetAnomalySpan(ctx core.Context, req *request.GetAnomalySpanRequest) (response.GetAnomalySpanResponse, error) {
	if req.PageParam == nil {
		req.PageParam = &request.PageParam{
			CurrentPage: 1,
			PageSize:    999,
		}
	}
	resp := response.GetAnomalySpanResponse{}
	results, total, err := s.chRepo.GetAnomalyTrace(ctx, req)
	if err != nil {
		return resp, err
	}

	resp.List = results
	resp.Pagination = &model.Pagination{
		Total:       total,
		CurrentPage: req.CurrentPage,
		PageSize:    req.PageSize,
	}

	return resp, nil
}
