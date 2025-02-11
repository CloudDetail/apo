// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetServiceInstanceOptions(req *request.GetServiceInstanceOptionsRequest) (map[string]*model.ServiceInstance, error) {
	// Get the list of active service instances
	instances, err := s.promRepo.GetActiveInstanceList(req.StartTime, req.EndTime, req.ServiceName, nil)
	if err != nil {
		return nil, err
	}

	return instances.GetInstanceIdMap(), nil
}
