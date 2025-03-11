// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryLog query full logs
// @Summary query all logs
// @Description query full logs
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.LogQueryRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogQueryResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/query [post]
func (h *handler) QueryLog() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogQueryRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}
		if req.PageNum == 0 {
			req.PageNum = 1
		}
		if req.PageSize == 0 {
			req.PageSize = 10
		}
		if req.Query == "" {
			req.Query = "(1='1')"
		}
		if req.TimeField == "" {
			req.TimeField = "timestamp"
		}
		if req.LogField == "" {
			req.LogField = "content"
		}
		resp, err := h.logService.QueryLog(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.QueryLogError,
				c.ErrMessage(code.QueryLogError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
