// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

func (s *service) GetTraceFilterValues(ctx core.Context, startTime, endTime time.Time, searchText string, filter request.SpanTraceFilter) (*response.GetTraceFilterValueResponse, error) {
	option, err := s.chRepo.GetFieldValues(ctx, searchText, &filter, startTime, endTime)
	if err != nil {
		return &response.GetTraceFilterValueResponse{
			TraceFilterOptions: clickhouse.SpanTraceOptions{
				SpanTraceFilter: filter,
				Options:         []string{},
			},
		}, err
	}

	return &response.GetTraceFilterValueResponse{
		TraceFilterOptions: *option,
	}, nil
}
