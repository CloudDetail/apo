// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// AddAlertRule new alarm rules
// @Summary new alarm rules
// @Description new alarm rules
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.AddAlertRuleRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/add [post]
func (h *handler) AddAlertRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AddAlertRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.alertService.AddAlertRule(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.AddAlertRuleError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
