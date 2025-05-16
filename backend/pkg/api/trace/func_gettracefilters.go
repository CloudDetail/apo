// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetTraceFilters the available filters for querying the Trace list
// @Summary the available filters for querying the Trace list
// @Description the available filters for querying the Trace list
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param needUpdate query bool false "Whether to update the available filters immediately based on the time entered by the user"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetTraceFiltersResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/pagelist/filters [get]
func (h *handler) GetTraceFilters() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTraceFiltersRequest)
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
		resp, err := h.traceService.GetTraceFilters(startTime, endTime, req.NeedUpdate)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetTraceFiltersError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
