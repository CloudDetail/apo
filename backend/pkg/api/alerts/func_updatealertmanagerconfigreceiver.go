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

// UpdateAlertManagerConfigReceiver update alarm notification object
// @Summary update alarm notification object
// @Description update alarm notification object
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.UpdateAlertManagerConfigReceiver true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alertmanager/receiver [post]
func (h *handler) UpdateAlertManagerConfigReceiver() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateAlertManagerConfigReceiver)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.UpdateAMConfigReceiver(req)
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
					code.UpdateAMConfigReceiverError,
					c.ErrMessage(code.UpdateAMConfigReceiverError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
