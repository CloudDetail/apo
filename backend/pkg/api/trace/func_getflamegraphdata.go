// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetFlameGraphData get the flame map data of the specified time period and specified conditions
// @Summary get the flame chart data of the specified time period and specified conditions
// @Description get the flame chart data of the specified time period and specified conditions
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param sampleType query string true "sample type"
// @Param pid query uint64 true "process id"
// @Param tid query uint64 true "thread id"
// @Param nodeName query string false "hostname"
// @Param startTime query int64 true "start time"
// @Param endTime query int64 true "end time"
// @Param spanId query string true "span id"
// @Param traceId query string true "trace id"
// @Success 200 {object} response.GetFlameDataResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/flame [get]
func (h *handler) GetFlameGraphData() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetFlameDataRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.traceService.GetFlameGraphData(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFlameGraphError,
				c.ErrMessage(code.GetFlameGraphError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
