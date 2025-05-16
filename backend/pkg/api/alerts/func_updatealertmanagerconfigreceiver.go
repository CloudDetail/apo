// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
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
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.alertService.UpdateAMConfigReceiver(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UpdateAMConfigReceiverError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
