// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetTracePageList to query the trace paging list
// @Summary to query the trace paging list
// @Description to query the trace paging list
// @Tags API.trace
// @Accept json
// @Produce json
// @Param Request body request.GetTracePageListRequest true "Request information"
// @Success 200 {object} response.GetTracePageListResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/pagelist [post]
func (h *handler) GetTracePageList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTracePageListRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		if req.PageNum == 0 {
			req.PageNum = 1
		}
		if req.PageSize == 0 {
			req.PageSize = 10
		}

		resp, err := h.traceService.GetTracePageList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetTracePageListError,
				code.Text(code.GetTracePageListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
