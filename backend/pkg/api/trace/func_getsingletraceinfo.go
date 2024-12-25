// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetSingleTraceInfo 获取单链路Trace详情
// @Summary 获取单链路Trace详情
// @Description 获取单链路Trace详情
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
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.traceService.GetSingleTraceID(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetSingleTraceError,
				code.Text(code.GetSingleTraceError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
