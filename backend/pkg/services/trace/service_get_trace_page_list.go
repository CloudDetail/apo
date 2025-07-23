// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) GetTracePageList(ctx core.Context, req *request.GetTracePageListRequest) (*response.GetTracePageListResponse, error) {
	if req.GroupID > 0 && len(req.Service) == 0 {
		selected, err := s.dbRepo.GetScopeIDsSelectedByGroupID(ctx, req.GroupID)
		if err != nil {
			return nil, err
		}
		permSvcList := common.DataGroupStorage.GetFullPermissionSvcList(selected)
		if len(permSvcList) == 0 {
			return &response.GetTracePageListResponse{
				Pagination: &model.Pagination{
					Total:       0,
					CurrentPage: req.PageNum,
					PageSize:    req.PageSize,
				},
				List: []clickhouse.QueryTraceResult{},
			}, nil
		}
		req.Service = permSvcList
	}

	for i := 0; i < len(req.Namespace); i++ {
		if req.Namespace[i] == "VM_NS" {
			req.Namespace[i] = ""
		}
	}

	for i := 0; i < len(req.Filters); i++ {
		if req.Filters[i].SpanTraceFilter == nil {
			continue
		}
		if req.Filters[i].Key == "namespace" && len(req.Filters[i].Value) > 0 {
			for s := 0; s < len(req.Filters[i].Value); s++ {
				if req.Filters[i].Value[s] == "VM_NS" {
					req.Filters[i].Value[s] = ""
				}
			}
		}
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
