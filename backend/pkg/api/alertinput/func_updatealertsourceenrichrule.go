// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// UpdateAlertSourceEnrichRule 更新告警源增强配置
// @Summary 更新告警源增强配置
// @Description 更新告警源增强配置
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.AlerEnrichRuleConfigRequest true "请求信息"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/enrich/update [post]
func (h *handler) UpdateAlertSourceEnrichRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.AlerEnrichRuleConfigRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.inputService.UpdateAlertEnrichRule(req)
		if err != nil {
			var vErr alert.ErrAlertSourceNotExist

			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AlertSourceNotExisted,
					code.Text(code.AlertSourceNotExisted)).WithError(err),
				)
				return
			}

			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UpdateAlertEnrichRuleFailed,
				code.Text(code.UpdateAlertEnrichRuleFailed)).WithError(err),
			)
			return
		}

		c.Payload("ok")
	}
}
