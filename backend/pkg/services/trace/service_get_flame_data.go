// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"go.uber.org/zap"
)

func (s *service) GetFlameGraphData(ctx_core core.Context, req *request.GetFlameDataRequest) (resp response.GetFlameDataResponse, err error) {
	flameData, err := s.chRepo.GetFlameGraphData(ctx_core, req.StartTime, req.EndTime, req.NodeName,
		req.PID, req.TID, req.SampleType, req.SpanID, req.TraceID)
	if err != nil {
		return
	}

	if len(*flameData) == 0 {
		return
	}
	if len(*flameData) > 1 {
		s.logger.Warn("invoke level flame graph should have one flame data", zap.Int("got", len(*flameData)))
	}
	resp = (response.GetFlameDataResponse)((*flameData)[0])
	return
}
