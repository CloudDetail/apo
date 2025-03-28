// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// UpdateAlertSourceEnrichRule UpdateAlertSourceEnrichRule
// @Summary UpdateAlertSourceEnrichRule
// @Description UpdateAlertSourceEnrichRule
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.AlertEnrichRuleConfigRequest true "Update Config"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/enrich/update [post]
func (h *handler) UpdateAlertSourceEnrichRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.AlertEnrichRuleConfigRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
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
					c.ErrMessage(code.AlertSourceNotExisted)).WithError(err),
				)
				return
			}

			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UpdateAlertEnrichRuleFailed,
				c.ErrMessage(code.UpdateAlertEnrichRuleFailed)).WithError(err),
			)
			return
		}

		c.Payload("ok")
	}
}
