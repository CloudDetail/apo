// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAlertRuleFile get basic alarm rules
// @Summary get basic alarm rules
// @Description get basic alarm rules
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param alertRuleFile query string false "Query alarm rule file name, if empty, return all"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetAlertRuleFileResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rules [get]
func (h *handler) GetAlertRuleFile() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertRuleConfigRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.alertService.GetAlertRuleFile(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertRuleError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
