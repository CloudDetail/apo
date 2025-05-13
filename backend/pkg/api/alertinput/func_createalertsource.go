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

// CreateAlertSource Create Alarm Source
// @Summary Create Alarm Source
// @Description Create Alarm Source
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.AlertSource true "AlertSource"
// @Success 200 {object} alert.AlertSource
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/create [post]
func (h *handler) CreateAlertSource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.AlertSource)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		alertSource, err := h.inputService.CreateAlertSource(req)
		if err != nil {
			var vErr alert.ErrAlertSourceAlreadyExist
			if errors.As(err, &vErr) {
				c.AbortWithError(
					http.StatusBadRequest,
					code.AlertSourceAlreadyExisted,
					err,
				)
				return
			}

			c.AbortWithError(
				http.StatusBadRequest,
				code.CreateAlertSourceFailed,
				err,
			)
			return
		}
		c.Payload(alertSource)
	}
}
