// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UpdateLogParseRule update log table parsing rules
// @Summary update log table parsing rules
// @Description update log table parsing rules
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.UpdateLogParseRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogParseResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/update [post]
func (h *handler) UpdateLogParseRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateLogParseRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.UpdateLogParseRule(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UpdateLogParseRuleError,
				c.ErrMessage(code.UpdateLogParseRuleError)+": "+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
