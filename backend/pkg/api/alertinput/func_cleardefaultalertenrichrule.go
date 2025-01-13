// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// ClearDefaultAlertEnrichRule 清除默认的告警丰富规则
// @Summary 清除默认的告警丰富规则
// @Description 清除默认的告警丰富规则
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.DefaultAlertEnrichRuleRequest true "请求信息"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/enrich/default/clear [get]
func (h *handler) ClearDefaultAlertEnrichRule() core.HandlerFunc {
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

		_, err := h.inputService.ClearDefaultAlertEnrichRule(req.SourceType)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ClearDefaultAlertEnrichRuleFailed,
				code.Text(code.ClearDefaultAlertEnrichRuleFailed)).WithError(err),
			)
		}
		c.Payload("ok")
	}
}
