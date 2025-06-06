// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryLogContext get the log context
// @Summary get log context
// @Description get log context
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.LogQueryContextRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogQueryContextResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/context [post]
func (h *handler) QueryLogContext() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogQueryContextRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp, err := h.logService.QueryLogContext(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.QueryLogContextError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
