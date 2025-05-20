// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UpdateAlertRule update alarm rules
// @Summary update alarm rules
// @Description update alarm rules
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.UpdateAlertRuleRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule [post]
func (h *handler) UpdateAlertRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateAlertRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.alertService.UpdateAlertRule(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UpdateAlertRuleError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
