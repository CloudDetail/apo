// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteAlertRule delete alarm rules
// @Summary delete alarm rules
// @Description delete alarm rules
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.DeleteAlertRuleRequest true "Delete object"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule [delete]
func (h *handler) DeleteAlertRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteAlertRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.DeleteAlertRule(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.DeleteAlertRuleError,
				code.Text(code.DeleteAlertRuleError)).WithError(err),
			)
			return
		}

		c.Payload("ok")
	}
}
