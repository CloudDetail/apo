// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAlertRules list alarm rules
// @Summary list alarm rules
// @Description list alarm rules
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.GetAlertRuleRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetAlertRulesResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/list [post]
func (h *handler) GetAlertRules() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp := h.alertService.GetAlertRules(req)
		c.Payload(resp)
	}
}
