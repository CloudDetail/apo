// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetTraceFilters(ctx core.Context, startTime, endTime time.Time, needUpdate bool) (*response.GetTraceFiltersResponse, error) {
	filters, err := s.chRepo.GetAvailableFilterKey(ctx, startTime, endTime, needUpdate)
	if err != nil {
		return &response.GetTraceFiltersResponse{
			TraceFilters: []request.SpanTraceFilter{},
		}, err
	}

	return &response.GetTraceFiltersResponse{
		TraceFilters: filters,
	}, nil
}
