// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UpdateAlertRuleFile update basic alarm rules
// @Summary update basic alarm rules
// @Description update basic alarm rules
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.UpdateAlertRuleRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rules/file [post]
func (h *handler) UpdateAlertRuleFile() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateAlertRuleConfigRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.UpdateAlertRuleFile(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UpdateAlertRuleError,
				code.Text(code.UpdateAlertRuleError)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
