// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// SetDefaultAlertEnrichRule SetDefaultAlertEnrichRule
// @Summary SetDefaultAlertEnrichRule
// @Description SetDefaultAlertEnrichRule
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.SetDefaultAlertEnrichRuleRequest true "Default EnrichRule"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/enrich/default/set [post]
func (h *handler) SetDefaultAlertEnrichRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.SetDefaultAlertEnrichRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.inputService.SetDefaultAlertEnrichRule(c, req.SourceType, req.EnrichRuleConfigs)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.SetDefaultAlertEnrichRuleFailed,
				err,
			)
		}
		c.Payload("ok")
	}
}
