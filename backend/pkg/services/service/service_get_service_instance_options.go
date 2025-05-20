// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetServiceInstanceOptions(ctx core.Context, req *request.GetServiceInstanceOptionsRequest) (map[string]*model.ServiceInstance, error) {
	// Get the list of active service instances
	instances, err := s.promRepo.GetActiveInstanceList(req.StartTime, req.EndTime, []string{req.ServiceName})
	if err != nil {
		return nil, err
	}

	return instances.GetInstanceIdMap(), nil
}
