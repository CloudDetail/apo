// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) ListAPPInfoLabelsKeys(ctx core.Context, req *request.QueryAPPInfoTagsRequest) (*response.QueryAPPInfoTagsResponse, error) {
	labels, err := s.chRepo.ListAppInfoLabelKeys(ctx, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	return &response.QueryAPPInfoTagsResponse{
		Labels: labels,
	}, nil
}

func (s *service) ListAPPInfoLabelValues(ctx core.Context, req *request.QueryAPPInfoTagValuesRequest) (*response.QueryAPPInfoTagValuesResponse, error) {
	values, err := s.chRepo.ListAppInfoLabelValues(ctx, req.StartTime, req.EndTime, req.Label)
	if err != nil {
		return nil, err
	}
	return &response.QueryAPPInfoTagValuesResponse{
		Labels: req.Label,
		Values: values,
	}, nil
}
