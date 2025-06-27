// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServiceList(ctx core.Context, req *request.GetServiceListRequest) ([]string, error) {
	filter := prometheus.NewFilter()
	if len(req.Namespace) > 0 {
		filter.RegexMatch("namespace", prometheus.RegexMultipleValue(req.Namespace...))
	}
	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch("cluster_id", prometheus.RegexMultipleValue(req.ClusterIDs...))
	}
	return s.promRepo.GetServiceList(ctx, req.StartTime, req.EndTime, filter)
}
