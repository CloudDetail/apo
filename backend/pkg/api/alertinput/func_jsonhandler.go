// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"fmt"
	"io"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	input "github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// JsonHandler 基于JSON结构接收来自特定数据源的数据
// @Summary 基于JSON结构接收来自特定数据源的数据
// @Description 基于JSON结构接收来自特定数据源的数据
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/event/json/:sourceType/:sourceName [post]
func (h *handler) JsonHandler() core.HandlerFunc {
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

		// 填充基础的SourceFrom信息
		if len(sourceFrom.SourceType) == 0 {
			sourceFrom.SourceType = "unkonwn"
		}
		if len(sourceFrom.SourceName) == 0 {
			sourceFrom.SourceName = fmt.Sprintf("%s-(%s)", sourceFrom.SourceType, c.ClientIP())
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
