// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// GetDefaultAlertEnrichRule 获取默认的告警丰富规则
// @Summary 获取默认的告警丰富规则
// @Description 获取默认的告警丰富规则
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.DefaultAlertEnrichRuleRequest true "请求信息"
// @Success 200 {object} alert.DefaultAlertEnrichRuleResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/enrich/default [get]
func (h *handler) GetDefaultAlertEnrichRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.DefaultAlertEnrichRuleRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
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
