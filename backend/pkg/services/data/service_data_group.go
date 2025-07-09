// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetDataGroup(ctx core.Context, req *request.GetDataGroupRequest) (resp response.GetDataGroupResponse, err error) {
	filter := model.DataGroupFilter{
		Name:           req.GroupName,
		PageSize:       &req.PageSize,
		CurrentPage:    &req.CurrentPage,
		DatasourceList: req.DataSourceList,
	}

	dataGroups, count, err := s.dbRepo.GetDataGroup(ctx, filter)
	if err != nil {
		return
	}

	resp.DataGroupList = dataGroups
	resp.Total = count
	resp.CurrentPage = req.CurrentPage
	resp.PageSize = req.PageSize
	return
}
