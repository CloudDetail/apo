// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServiceInstanceOptions(ctx core.Context, req *request.GetServiceInstanceOptionsRequest) (map[string]*model.ServiceInstance, error) {
	// Get the list of active service instances
	filter := prometheus.NewFilter()
	filter.Equal(prometheus.ServiceNameKey, req.ServiceName)
	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch("cluster_id", prometheus.RegexMultipleValue(req.ClusterIDs...))
	}
	instances, err := s.promRepo.GetInstanceListByPQLFilter(ctx, req.StartTime, req.EndTime, filter)
	if err != nil {
		return nil, err
	}

	return instances.GetInstanceIdMap(), nil
}
