// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// GetAlertSourceEnrichRule GetAlertSourceEnrichRule
// @Summary GetAlertSourceEnrichRule
// @Description GetAlertSourceEnrichRule
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.SourceFrom true "Source Info"
// @Success 200 {object} alert.GetAlertEnrichRuleResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/enrich/get [get]
func (h *handler) GetAlertSourceEnrichRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.SourceFrom)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		rules, err := h.inputService.GetAlertEnrichRule(req.SourceID)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertEnrichRuleFailed,
				c.ErrMessage(code.GetAlertEnrichRuleFailed)).WithError(err),
			)
			return
		}
		c.Payload(alert.GetAlertEnrichRuleResponse{
			SourceId:          req.SourceID,
			EnrichRuleConfigs: rules,
		})
	}
}
