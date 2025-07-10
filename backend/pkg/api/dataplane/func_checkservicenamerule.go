// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// CheckServiceNameRule Check servicename rule.
// @Summary Check servicename rule.
// @Description Check servicename rule.
// @Tags API.dataplane
// @Accept json
// @Produce json
// @Param Request body request.SetServiceNameRuleRequest true "Check ServiceName Rule Request"
// @Success 200 {object} response.CheckServiceNameRuleResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/servicename/checkRule [post]
func (h *handler) CheckServiceNameRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetServiceNameRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.dataplaneService.CheckServiceNameRule(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CheckServiceNameRuleError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
