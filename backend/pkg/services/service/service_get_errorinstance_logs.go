// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

func (s *service) GetErrorInstanceLogs(ctx core.Context, req *request.GetErrorInstanceLogsRequest) ([]clickhouse.FaultLogResult, error) {
	// Get the error instance fault site log
	query := &clickhouse.FaultLogQuery{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Service:   req.Service,
		Instance:  req.Instance,
		EndPoint:  req.Endpoint,
		Type:      1, // Error Only
		PageNum:   1,
		PageSize:  5,
	}
	list, _, err := s.chRepo.GetFaultLogPageList(query)
	if err != nil {
		return nil, err
	}
	return list, nil
}
