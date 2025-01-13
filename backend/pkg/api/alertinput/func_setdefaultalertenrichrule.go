// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// SetDefaultAlertEnrichRule 设置默认的告警丰富规则
// @Summary 设置默认的告警丰富规则
// @Description 设置默认的告警丰富规则
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.SetDefaultAlertEnrichRuleRequest true "请求信息"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/enrich/default/set [post]
func (h *handler) SetDefaultAlertEnrichRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.SetDefaultAlertEnrichRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.inputService.SetDefaultAlertEnrichRule(req.SourceType, req.EnrichRuleConfigs)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.SetDefaultAlertEnrichRuleFailed,
				code.Text(code.SetDefaultAlertEnrichRuleFailed)).WithError(err),
			)
		}
		c.Payload("ok")
	}
}
