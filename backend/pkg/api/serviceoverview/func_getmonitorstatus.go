// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetMonitorStatus get the service status monitored by kuma
// @Summary get the service status monitored by kuma
// @Description get the service status monitored by kuma
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetMonitorStatusResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/monitor/status [get]
func (h *handler) GetMonitorStatus() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetMonitorStatusRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)

		resp, err := h.serviceoverview.GetMonitorStatus(startTime, endTime)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetMonitorStatusError,
				err,
			)
		}
		c.Payload(resp)
	}
}
