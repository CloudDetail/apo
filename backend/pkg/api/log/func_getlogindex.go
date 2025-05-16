// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetLogIndex analysis field index
// @Summary analysis field index
// @Description analysis field index
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.LogIndexRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogIndexResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/index [post]
func (h *handler) GetLogIndex() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogIndexRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
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
		resp, err := h.logService.GetLogIndex(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetLogIndexError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
