// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// ListServiceNameRule List servicename rule.
// @Summary List servicename rule.
// @Description List servicename rule.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param serviceName query string true "Service name"
// @Param startTime query int64 true "Start time"
// @Param endTime query int64 true "End time"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.ListServiceNameRuleResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/servicename/listRule [get]
func (h *handler) ListServiceNameRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ListServiceNameRuleRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp, err := h.dataplaneService.ListServiceNameRule(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ListServiceNameRuleError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
