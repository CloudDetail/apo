// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetServiceList(ctx core.Context, req *request.GetServiceListRequest) ([]string, error) {
	return s.promRepo.GetServiceList(ctx, req.StartTime, req.EndTime, req.Namespace)
}
