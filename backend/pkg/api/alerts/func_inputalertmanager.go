// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"io"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// InputAlertManager get AlertManager alarm events
// @Summary get AlertManager alarm events
// @Description get AlertManager alarm events
// @Tags API.alerts
// @Accept application/json
// @Produce json
// @Param Request body request.InputAlertManagerRequest true "Request information"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/inputs/alertmanager [post]
func (h *handler) InputAlertManager() core.HandlerFunc {
	return func(c core.Context) {
		data, err := io.ReadAll(c.Request().Body)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.AcceptAlertEventFailed,
				err,
			)
			return
		}

		// using APO-VM-ALERT as default source
		sourceFrom := alert.SourceFrom{SourceID: alert.ApoVMAlertSourceID}
		err = h.inputService.ProcessAlertEvents(c, sourceFrom, data)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ProcessAlertEventFailed,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
