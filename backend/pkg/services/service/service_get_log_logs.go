// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

func (s *service) GetLogLogs(ctx core.Context, req *request.GetLogLogsRequest) ([]clickhouse.FaultLogResult, error) {
	// Get Log fault field log
	query := &clickhouse.FaultLogQuery{
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Service:     req.Service,
		Instance:    req.Instance,
		NodeName:    req.NodeName,
		ContainerId: req.ContainerId,
		Pid:         req.Pid,
		EndPoint:    req.Endpoint,
		Type:        0, // Slow && Error & Profiled
		PageNum:     1,
		PageSize:    5,
		ClusterIDs:  req.ClusterIDs,
	}
	list, _, err := s.chRepo.GetFaultLogPageList(ctx, query)
	if err != nil {
		return nil, err
	}
	return list, nil
}
