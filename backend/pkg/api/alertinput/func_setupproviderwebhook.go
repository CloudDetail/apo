// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// SetupAlertProviderWebhook Install or update webhook
// @Summary Install or update webhook
// @Description Install or update webhook
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.SetupAlertProviderWebhookRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/webhook [post]
func (h *handler) SetupAlertProviderWebhook() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetupAlertProviderWebhookRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.inputService.SetupProviderWebhook(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				code.SetupAlertProviderWebhookError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
