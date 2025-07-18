// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) GetTracePageList(ctx core.Context, req *request.GetTracePageListRequest) (*response.GetTracePageListResponse, error) {
	if req.GroupID > 0 && len(req.Service) == 0 {
		selected, err := s.dbRepo.GetScopeIDsSelectedByGroupID(ctx, req.GroupID)
		if err != nil {
			return nil, err
		}
		permSvcList := common.DataGroupStorage.GetFullPermissionSvcList(selected)
		req.Service = permSvcList
	}

	list, total, err := s.chRepo.GetTracePageList(ctx, req)
	if err != nil {
		return nil, err
	}
	return &response.GetTracePageListResponse{
		Pagination: &model.Pagination{
			Total:       total,
			CurrentPage: req.PageNum,
			PageSize:    req.PageSize,
		},
		List: list,
	}, nil
}
