// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// MarkAlertResolvedManually
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.MarkAlertResolvedManuallyRequest true "请求信息"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/alerts/resolve [post]
func (h *handler) MarkAlertResolvedManually() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.MarkAlertResolvedManuallyRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.alertService.ManualResolveLatestAlertEventByAlertID(c, req.AlertID)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertSlienceError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
