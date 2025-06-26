// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetTraceFilters(ctx core.Context, req *request.GetTraceFiltersRequest) (*response.GetTraceFiltersResponse, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)
	filters, err := s.chRepo.GetAvailableFilterKey(ctx, startTime, endTime, req.NeedUpdate)
	if err != nil {
		return &response.GetTraceFiltersResponse{
			TraceFilters: []request.SpanTraceFilter{},
		}, err
	}

	return &response.GetTraceFiltersResponse{
		TraceFilters: filters,
	}, nil
}
