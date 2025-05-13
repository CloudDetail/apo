// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) GetServiceList(ctx_core core.Context, req *request.GetServiceListRequest) ([]string, error) {
	return s.promRepo.GetServiceList(req.StartTime, req.EndTime, req.Namespace)
}
