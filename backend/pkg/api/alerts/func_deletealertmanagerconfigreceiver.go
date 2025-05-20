// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteAlertManagerConfigReceiver delete alarm notification object
// @Summary delete alarm notification object
// @Description delete alarm notification object
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.DeleteAlertManagerConfigReceiverRequest true "Delete object"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alertmanager/receiver [delete]
func (h *handler) DeleteAlertManagerConfigReceiver() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteAlertManagerConfigReceiverRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.alertService.DeleteAMConfigReceiver(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.DeleteAlertRuleError,
				err,
			)
			return
		}

		c.Payload("ok")
	}
}
