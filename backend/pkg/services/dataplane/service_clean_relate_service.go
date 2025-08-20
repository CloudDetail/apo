// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) CleanRelateServiceByRuleId(ctx core.Context, req *request.CleanRuleServicesRequest) error {
	// return s.chRepo.CleanDataplaneServiceInstance(ctx, req.StartTime, req.EndTime, req.RuleId, "")
	return nil
}

func (s *service) CleanRelateServiceByDataSource(ctx core.Context, req *request.CleanDataSourceServicesRequest) error {
	return nil
	// return s.chRepo.CleanDataplaneServiceInstance(ctx, req.StartTime, req.EndTime, 0, req.DataSource)
}
