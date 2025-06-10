// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// ListAlertSlienceConfig
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} response.ListAlertSlienceConfigResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/slient/list [get]
func (h *handler) ListAlertSlienceConfig() core.HandlerFunc {
	return func(c core.Context) {
		sliences, err := h.alertService.ListSlienceConfig(c)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertSlienceError,
				err,
			)
			return
		}
		c.Payload(response.ListAlertSlienceConfigResponse{
			Sliences: sliences,
		})
	}
}
