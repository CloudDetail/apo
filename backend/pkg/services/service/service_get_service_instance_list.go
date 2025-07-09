// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"strconv"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServiceInstanceList(ctx core.Context, req *request.GetServiceInstanceListRequest) ([]string, error) {
	// Get the list of active service instances
	filter := prometheus.NewFilter()
	filter.Equal(prometheus.ServiceNameKey, req.ServiceName)
	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch(prometheus.ClusterIDKey, prometheus.RegexMultipleValue(req.ClusterIDs...))
	}
	instances, err := s.promRepo.GetInstanceListByPQLFilter(ctx, req.StartTime, req.EndTime, filter)
	if err != nil {
		return nil, err
	}
	return instances.GetInstanceIds(), nil
}

func (s *service) GetServiceInstanceInfoList(ctx core.Context, req *request.GetServiceInstanceListRequest) ([]prometheus.InstanceKey, error) {
	filter := prometheus.NewFilter()
	filter.Equal(prometheus.ServiceNameKey, req.ServiceName)
	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch(prometheus.ClusterIDKey, prometheus.RegexMultipleValue(req.ClusterIDs...))
	}
	instanceList, err := s.promRepo.GetInstanceListByPQLFilter(ctx, req.StartTime, req.EndTime, filter)
	if err != nil {
		return []prometheus.InstanceKey{}, err
	}

	// Fill the instance
	var ins []prometheus.InstanceKey
	for _, instance := range instanceList.GetInstanceIdMap() {
		key := prometheus.InstanceKey{
			PID:         strconv.FormatInt(instance.Pid, 10),
			ContainerId: instance.ContainerId,
			Pod:         instance.PodName,
			Namespace:   instance.Namespace,
			NodeName:    instance.NodeName,
			NodeIP:      instance.NodeIP,
		}
		ins = append(ins, key)
	}
	return ins, nil
}
