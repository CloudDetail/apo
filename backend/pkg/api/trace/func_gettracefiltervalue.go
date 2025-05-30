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

// GetTraceFilterValue query the available values of the specified filter
// @Summary query the available values of the specified filter
// @Description query the available values of the specified filter
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetTraceFilterValueRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetTraceFilterValueResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/pagelist/filter/value [post]
func (h *handler) GetTraceFilterValue() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTraceFilterValueRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)
		resp, err := h.traceService.GetTraceFilterValues(c, startTime, endTime, req.SearchText, req.Filter)

		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetTraceFilterValueError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
