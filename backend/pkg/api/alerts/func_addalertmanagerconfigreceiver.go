// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AddAlertManagerConfigReceiver new alarm notification object
// @Summary new alarm notification object
// @Description new alarm notification object
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.AddAlertManagerConfigReceiver true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alertmanager/receiver/add [post]
func (h *handler) AddAlertManagerConfigReceiver() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AddAlertManagerConfigReceiver)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.AddAMConfigReceiver(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					c.ErrMessage(vErr.Code),
				).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AddAMConfigReceiverError,
					c.ErrMessage(code.AddAMConfigReceiverError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
