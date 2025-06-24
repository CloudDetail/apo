// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetAlertEventReport(ctx core.Context, req *request.GetAlertEventReportRequest) (*response.GetAlertEventReportResponse, error) {

	report, err := s.chRepo.GetAlertEventReport(ctx, time.UnixMicro(req.StartTime), time.UnixMicro(req.EndTime+1000000), req.Tags)
	if err != nil {
		return nil, err
	}
	resp := response.ToAlertReportResponse(*report)
	return &resp, nil
}
