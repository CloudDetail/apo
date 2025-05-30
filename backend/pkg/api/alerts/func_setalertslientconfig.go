// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// SetAlertSlienceConfig
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.SetAlertSlienceConfigRequest true "请求信息"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/alerts/slient [post]
func (h *handler) SetAlertSlienceConfig() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetAlertSlienceConfigRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.alertService.SetSlienceConfigByAlertID(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.SetAlertSlienceError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
