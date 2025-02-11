// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"strconv"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServiceInstanceList(req *request.GetServiceInstanceListRequest) ([]string, error) {
	// Get the list of active service instances
	instances, err := s.promRepo.GetActiveInstanceList(req.StartTime, req.EndTime,[]string{req.ServiceName})
	if err != nil {
		return nil, err
	}

	return instances.GetInstanceIds(), nil
}

func (s *service) GetServiceInstanceInfoList(req *request.GetServiceInstanceListRequest) ([]prometheus.InstanceKey, error) {
	var ins []prometheus.InstanceKey
	// Get instance
	instanceList, err := s.promRepo.GetInstanceList(req.StartTime, req.EndTime, req.ServiceName, "")
	if err != nil {
		return ins, err
	}

	// Fill the instance
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
