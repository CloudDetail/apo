// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// DeleteAlertSource DeleteAlertSource
// @Summary DeleteAlertSource
// @Description DeleteAlertSource
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.SourceFrom true "alertSource"
// @Success 200 {object} alert.SourceFrom
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/delete [post]
func (h *handler) DeleteAlertSource() core.HandlerFunc {
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

		alert, err := h.inputService.DeleteAlertSource(*req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.DeleteAlertSourceFailed,
				c.ErrMessage(code.DeleteAlertSourceFailed)).WithError(err),
			)
			return

		}
		c.Payload(alert)
	}
}
