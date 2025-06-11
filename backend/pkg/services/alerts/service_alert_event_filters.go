package alerts

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetStaticFilterKeys(ctx core.Context) *response.AlertEventFiltersResponse {
	filters := s.chRepo.GetStaticFilterKeys(ctx)
	return &response.AlertEventFiltersResponse{Filters: filters}
}

func (s *service) GetAlertEventFilterLabelKeys(
	ctx core.Context,
	req *request.SearchAlertEventFilterValuesRequest,
) (*response.AlertEventFilterLabelKeysResponse, error) {
	labels, err := s.chRepo.GetAlertEventFilterLabelKeys(ctx, req)
	if err != nil {
		return nil, err
	}
	return &response.AlertEventFilterLabelKeysResponse{Labels: labels}, nil
}
func (s *service) GetAlertEventFilterValues(
	ctx core.Context,
	req *request.SearchAlertEventFilterValuesRequest,
) (*request.AlertEventFilter, error) {
	return s.chRepo.GetAlertEventFilterValues(ctx, req)
}
