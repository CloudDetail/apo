// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"fmt"
	"io"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	input "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// JsonHandler Receive data from a specific data source based on a JSON structure
// @Summary Receive data from a specific data source based on a JSON structure
// @Description Receive data from a specific data source based on a JSON structure
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/event/json [post]
func (h *handler) JsonHandler() core.HandlerFunc {
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

		// Fill the SourceFrom information of the base
		if len(sourceFrom.SourceType) == 0 {
			sourceFrom.SourceType = "unkonwn"
		}
		if len(sourceFrom.SourceName) == 0 {
			sourceFrom.SourceName = fmt.Sprintf("%s-(%s)", sourceFrom.SourceType, c.ClientIP())
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
