// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// ListAlertSource ListAlertSource
// @Summary ListAlertSource
// @Description ListAlertSource
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} alert.ListAlertSourceResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/list [get]
func (h *handler) ListAlertSource() core.HandlerFunc {
	return func(c core.Context) {
		alertSources, err := h.inputService.ListAlertSource()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ListAlertSourceFailed,
				code.Text(code.ListAlertSourceFailed)).WithError(err),
			)
			return
		}

		c.Payload(alert.ListAlertSourceResponse{
			AlertSources: alertSources,
		})
	}
}
