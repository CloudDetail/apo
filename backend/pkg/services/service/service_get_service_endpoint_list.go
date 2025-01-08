// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import "github.com/CloudDetail/apo/backend/pkg/model/request"

func (s *service) GetServiceEndPointList(req *request.GetServiceEndPointListRequest) ([]string, error) {
	// Get the list of service Endpoint
	return s.promRepo.GetServiceEndPointList(req.StartTime, req.EndTime, req.ServiceName)
}
