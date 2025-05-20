// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetOnOffCPU(ctx core.Context, req *request.GetOnOffCPURequest) (*response.GetOnOffCPUResponse, error) {
	result, err := s.chRepo.GetOnOffCPU(req.PID, req.NodeName, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	resp := &response.GetOnOffCPUResponse{
		ProfilingEvent: result,
	}
	return resp, nil
}
