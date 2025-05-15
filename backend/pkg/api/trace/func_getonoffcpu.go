// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetOnOffCPU get span execution consumption
// @Summary get span execution consumption
// @Description get span execution consumption
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "start time"
// @Param endTime query int64 true "end time"
// @Param pid query uint32 true "process id"
// @Param nodeName query string true "node name"
// @Success 200 {object} response.GetOnOffCPUResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/onoffcpu [get]
func (h *handler) GetOnOffCPU() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetOnOffCPURequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.traceService.GetOnOffCPU(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetOnOffCPUError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
