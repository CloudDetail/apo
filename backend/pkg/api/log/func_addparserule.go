// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AddLogParseRule new log table parsing rules
// @Summary new log table parsing rules
// @Description new log table parsing rules
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.AddLogParseRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogParseResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/add [post]
func (h *handler) AddLogParseRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AddLogParseRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp, err := h.logService.AddLogParseRule(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.AddLogParseRuleError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
