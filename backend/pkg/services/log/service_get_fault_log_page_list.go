// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) GetFaultLogPageList(ctx core.Context, req *request.GetFaultLogPageListRequest) (*response.GetFaultLogPageListResponse, error) {
	// Paging query fault site logs
	query := &clickhouse.FaultLogQuery{
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		MultiServices:  req.Service,
		MultiNamespace: req.Namespaces,
		NodeName:       req.NodeName,
		ContainerId:    req.ContainerId,
		Pid:            req.Pid,
		Instance:       req.Instance,
		TraceId:        req.TraceId,
		Type:           2, // Slow && Error && Normal
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		Pod:            req.Pod,
		ClusterIDs:     req.ClusterIDs,
	}

	if req.GroupID > 0 && len(query.MultiServices) == 0 {
		selected, err := s.dbRepo.GetScopeIDsSelectedByGroupID(ctx, req.GroupID)
		if err != nil {
			return nil, err
		}
		permSvcList := common.DataGroupStorage.GetFullPermissionSvcList(selected)
		if len(permSvcList) == 0 {
			return &response.GetFaultLogPageListResponse{
				Pagination: &model.Pagination{
					Total:       0,
					CurrentPage: req.PageNum,
					PageSize:    req.PageSize,
				},
				List: []clickhouse.FaultLogResult{},
			}, nil
		}
		query.MultiServices = permSvcList
	}

	list, total, err := s.chRepo.GetFaultLogPageList(ctx, query)
	if err != nil {
		return nil, err
	}
	return &response.GetFaultLogPageListResponse{
		Pagination: &model.Pagination{
			Total:       total,
			CurrentPage: req.PageNum,
			PageSize:    req.PageSize,
		},
		List: list,
	}, nil
}
