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

// UpdateAlertSource UpdateAlertSource
// @Summary UpdateAlertSource
// @Description UpdateAlertSource
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.AlertSource true "alertSource Info"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/update [post]
func (h *handler) UpdateAlertSource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.AlertSource)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		source, err := h.inputService.UpdateAlertSource(req)
		if err != nil {
			var vErr alert.ErrAlertSourceAlreadyExist
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AlertSourceAlreadyExisted,
					code.Text(code.AlertSourceAlreadyExisted)).WithError(err),
				)
				return
			}

			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UpdateAlertSourceFailed,
				code.Text(code.UpdateAlertSourceFailed)).WithError(err),
			)
			return
		}
		c.Payload(source)
	}
}
