// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServiceNamespaceList(ctx core.Context, req *request.GetServiceNamespaceListRequest) (response.GetServiceNamespaceListResponse, error) {
	filter := prometheus.NewFilter()
	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch(prometheus.ClusterIDKey, prometheus.RegexMultipleValue(req.ClusterIDs...))
	}
	list, err := s.promRepo.GetNamespaceList(ctx, req.StartTime, req.EndTime, filter)
	var resp response.GetServiceNamespaceListResponse
	if err != nil {
		return resp, err
	}

	resp.NamespaceList = list
	return resp, nil
}
