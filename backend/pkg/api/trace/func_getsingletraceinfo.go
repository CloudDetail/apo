// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetSingleTraceInfo get single-link Trace details
// @Summary get single link trace details
// @Description get single link trace details
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param traceId query string true "trace id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetSingleTraceInfoResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/info [get]
func (h *handler) GetSingleTraceInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSingleTraceInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.traceService.GetSingleTraceID(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetSingleTraceError,
				c.ErrMessage(code.GetSingleTraceError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
