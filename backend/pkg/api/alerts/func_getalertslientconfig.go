// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetAlertSlienceConfig
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetAlertSlienceConfigRequest true "请求信息"
// @Success 200 {object} response.GetAlertSlienceConfigResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/slient [get]
func (h *handler) GetAlertSlienceConfig() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertSlienceConfigRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		config, err := h.alertService.GetSlienceConfigByAlertID(req.AlertID)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertSlienceError,
				err,
			)
			return
		}
		c.Payload(response.GetAlertSlienceConfigResponse{
			Slience: config,
		})
	}
}
