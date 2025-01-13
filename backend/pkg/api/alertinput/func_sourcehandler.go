// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"io"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	input "github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// SourceHandler 基于告警源配置接收数据
// @Summary 基于告警源配置接收数据
// @Description 基于告警源配置接收数据
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/event/source/:sourceID [post]
func (h *handler) SourceHandler() core.HandlerFunc {
	return func(c core.Context) {
		var sourceFrom input.SourceFrom
		err := c.ShouldBindURI(&sourceFrom)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		data, err := io.ReadAll(c.Request().Body)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.AcceptAlertEventFailed,
				code.Text(code.AcceptAlertEventFailed)).WithError(err),
			)
			return
		}
		err = h.inputService.ProcessAlertEvents(sourceFrom, data)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ProcessAlertEventFailed,
				code.Text(code.ProcessAlertEventFailed)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
