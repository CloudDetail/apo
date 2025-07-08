// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServiceEndPointList(ctx core.Context, req *request.GetServiceEndPointListRequest) ([]string, error) {
	// Get the list of service Endpoint
	filter := prometheus.NewFilter()
	filter.Equal(prometheus.ServiceNameKey, req.ServiceName)
	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch(prometheus.ClusterIDKey, prometheus.RegexMultipleValue(req.ClusterIDs...))
	}
	return s.promRepo.GetServiceEndPointListByPQLFilter(ctx, req.StartTime, req.EndTime, filter)
}
