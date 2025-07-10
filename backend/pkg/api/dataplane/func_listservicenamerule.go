// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// ListServiceNameRule List servicename rule.
// @Summary List servicename rule.
// @Description List servicename rule.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.ListServiceNameRuleResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/servicename/listRule [get]
func (h *handler) ListServiceNameRule() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.dataplaneService.ListServiceNameRule(c)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ListServiceNameRuleError,
				err,
			)
		}
		c.Payload(resp)
	}
}
