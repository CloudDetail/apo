// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// GetAlertSource 获取告警源信息
// @Summary 获取告警源信息
// @Description 获取告警源信息
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.SourceFrom true "告警源信息"
// @Success 200 {object} alert.AlertSource
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/get [post]
func (h *handler) GetAlertSource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.SourceFrom)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.inputService.GetAlertSource(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.CreateAlertSourceFailed,
				// TODO ErrorCode
				code.Text(code.GetAlertSourceFailed)).WithError(err),
			)
			return
		}

		c.Payload(resp)
	}
}
