// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetLogParseRule get log table parsing rules
// @Summary get log table parsing rules
// @Description get log table parsing rules
// @Tags API.log
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param dataBase query string true "database"
// @Param tableName query string true "Table"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogParseResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/get [get]
func (h *handler) GetLogParseRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.QueryLogParseRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp, err := h.logService.GetLogParseRule(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetLogParseRuleError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
