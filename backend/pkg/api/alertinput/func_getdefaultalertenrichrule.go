// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// GetDefaultAlertEnrichRule GetDefaultAlertEnrichRule
// @Summary GetDefaultAlertEnrichRule
// @Description GetDefaultAlertEnrichRule
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.DefaultAlertEnrichRuleRequest true "Source Type"
// @Success 200 {object} alert.DefaultAlertEnrichRuleResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/enrich/default [get]
func (h *handler) GetDefaultAlertEnrichRule() core.HandlerFunc {
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

		sourceType, rules := h.inputService.GetDefaultAlertEnrichRule(req.SourceType)
		c.Payload(alert.DefaultAlertEnrichRuleResponse{
			SourceType:        sourceType,
			EnrichRuleConfigs: rules,
		})
	}
}
