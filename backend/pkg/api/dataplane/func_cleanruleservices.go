// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// CleanRuleServices clean relate service by ruleId.
// @Summary clean relate service by ruleId.
// @Description clean relate service by ruleId.
// @Tags API.dataplane
// @Accept json
// @Produce json
// @Param Request body request.CleanRuleServicesRequest true "Request information"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/servicename/cleanRuleServices [post]
func (h *handler) CleanRuleServices() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CleanRuleServicesRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if err := h.dataplaneService.CleanRelateServiceByRuleId(c, req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CleanRelateServiceError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
