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
// @Accept json
// @Produce json
// @Param Request body request.GetAlertEventReportRequest true "请求信息"
// @Success 200 {object} response.GetAlertEventReportResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/events/report/get [post]
func (h *handler) GetAlertEventReport() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertEventReportRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.alertService.GetAlertEventReport(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertEventReportError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
