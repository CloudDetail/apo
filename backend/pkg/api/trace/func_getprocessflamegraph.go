// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetProcessFlameGraph capture and integrate process-level flame graph data
// @Summary get and integrate process-level flame graph data
// @Description get and integrate process-level flame graph data
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param maxNodes query int64 false "Limit number of nodes"
// @Param startTime query int64 true "start time"
// @Param endTime query int64 true "end time"
// @Param pid query int64 true "process id"
// @Param nodeName query string false "hostname"
// @Param sampleType query string true "sample type"
// @Success 200 {object} response.GetProcessFlameGraphResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/flame/process [get]
func (h *handler) GetProcessFlameGraph() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetProcessFlameGraphRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.traceService.GetProcessFlameGraphData(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFlameGraphError,
				code.Text(code.GetFlameGraphError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
