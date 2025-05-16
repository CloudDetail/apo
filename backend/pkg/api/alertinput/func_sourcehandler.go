// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"io"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	input "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// SourceHandler Receive data based on alarm source configuration
// @Summary Receive data based on alarm source configuration
// @Description Receive data based on alarm source configuration
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/event/source [post]
func (h *handler) SourceHandler() core.HandlerFunc {
	return func(c core.Context) {
		var sourceFrom input.SourceFrom
		err := c.ShouldBindQuery(&sourceFrom)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		data, err := io.ReadAll(c.Request().Body)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.AcceptAlertEventFailed,
				err,
			)
			return
		}
		err = h.inputService.ProcessAlertEvents(sourceFrom, data)
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
