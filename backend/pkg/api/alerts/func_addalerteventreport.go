// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AlertEventList
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.AlertEventClassifyRequest true "请求信息"
// @Success 200 {object} response.AlertEventClassifyResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/events/classify [get]
func (h *handler) AddAlertEventReport() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AddAlertEventReportRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.alertService.AddAlertEventReport(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.AddAlertEventReportError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
