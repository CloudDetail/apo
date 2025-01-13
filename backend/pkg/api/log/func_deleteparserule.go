// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteLogParseRule delete log table parsing rules
// @Summary delete log table parsing rules
// @Description delete log table parsing rules
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.DeleteLogParseRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogParseResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/delete [delete]
func (h *handler) DeleteLogParseRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteLogParseRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.DeleteLogParseRule(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.DeleteLogParseRuleError,
				code.Text(code.DeleteLogParseRuleError)+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
