// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) AddAlertEventReport(ctx core.Context, req *request.AddAlertEventReportRequest) error {
	report, _ := s.chRepo.GetAlertEventReport(ctx, time.UnixMicro(req.Overview.TimeStamp-3600000000), time.UnixMicro(req.Overview.TimeStamp+3600000000), req.Tags)
	if report != nil {
		return core.Error(code.AddAlertEventReportError, "alert event report already exists")
	}
	err := s.chRepo.AddAlertEventReport(ctx, req)
	return err
}
