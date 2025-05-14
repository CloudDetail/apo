// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) GetFaultLogPageList(ctx_core core.Context, req *request.GetFaultLogPageListRequest) (*response.GetFaultLogPageListResponse, error) {
	// Paging query fault site logs
	query := &clickhouse.FaultLogQuery{
		StartTime:	req.StartTime,
		EndTime:	req.EndTime,
		MultiServices:	req.Service,
		MultiNamespace:	req.Namespaces,
		NodeName:	req.NodeName,
		ContainerId:	req.ContainerId,
		Pid:		req.Pid,
		Instance:	req.Instance,
		TraceId:	req.TraceId,
		Type:		2,	// Slow && Error && Normal
		PageNum:	req.PageNum,
		PageSize:	req.PageSize,
		Pod:		req.Pod,
	}
	list, total, err := s.chRepo.GetFaultLogPageList(ctx_core, query)
	if err != nil {
		return nil, err
	}
	return &response.GetFaultLogPageListResponse{
		Pagination: &model.Pagination{
			Total:		total,
			CurrentPage:	req.PageNum,
			PageSize:	req.PageSize,
		},
		List:	list,
	}, nil
}
