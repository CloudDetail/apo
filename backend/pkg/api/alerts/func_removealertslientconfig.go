// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// RemoveAlertSlienceConfig
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.RemoveAlertSlienceConfigRequest true "请求信息"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/alerts/slient [delete]
func (h *handler) RemoveAlertSlienceConfig() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.RemoveAlertSlienceConfigRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		err := h.alertService.RemoveSlienceConfigByAlertID(c, req.AlertID)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.RemoveAlertSlienceError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
