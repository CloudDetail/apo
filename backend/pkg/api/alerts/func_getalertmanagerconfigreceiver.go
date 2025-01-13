// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAlertManagerConfigReceiver list alarm notification objects
// @Summary list alarm notification objects
// @Description list alarm notification objects
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.GetAlertManagerConfigReceverRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetAlertManagerConfigReceiverResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alertmanager/receiver/list [post]
func (h *handler) GetAlertManagerConfigReceiver() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertManagerConfigReceverRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp := h.alertService.GetAMConfigReceivers(req)
		c.Payload(resp)
	}
}
