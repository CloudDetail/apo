// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) GetServiceEndPointList(ctx_core core.Context, req *request.GetServiceEndPointListRequest) ([]string, error) {
	// Get the list of service Endpoint
	return s.promRepo.GetServiceEndPointList(req.StartTime, req.EndTime, req.ServiceName)
}
