// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// ClearDefaultAlertEnrichRule Clear default alarm rich rules
// @Summary Clear default alarm rich rules
// @Description Clear default alarm rich rules
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.DefaultAlertEnrichRuleRequest true "Request info"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/enrich/default/clear [get]
func (h *handler) ClearDefaultAlertEnrichRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.DefaultAlertEnrichRuleRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		_, err := h.inputService.ClearDefaultAlertEnrichRule(c, req.SourceType)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ClearDefaultAlertEnrichRuleFailed,
				err,
			)
		}
		c.Payload("ok")
	}
}
