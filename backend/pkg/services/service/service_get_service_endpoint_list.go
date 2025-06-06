// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetServiceEndPointList(ctx core.Context, req *request.GetServiceEndPointListRequest) ([]string, error) {
	// Get the list of service Endpoint
	return s.promRepo.GetServiceEndPointList(ctx, req.StartTime, req.EndTime, req.ServiceName)
}
