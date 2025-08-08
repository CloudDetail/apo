// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetDatasources(ctx core.Context, req *request.GetDatasourcesRequest) (response.GetDatasourcesResponse, error) {
	var resp response.GetDatasourcesResponse
	datasourceCount, err := s.dbRepo.GetDatapalenDsCount(ctx)
	if err != nil {
		return resp, err
	}

	count, err := s.chRepo.QueryServiceCount(ctx, req.StartTime, req.EndTime, req.Cluster)
	if err != nil {
		return resp, err
	}

	resp.ServiceCount = count
	resp.Datasources = datasourceCount
	return resp, nil
}
