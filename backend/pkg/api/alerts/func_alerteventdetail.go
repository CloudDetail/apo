// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AlertEventDetail
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetAlertDetailRequest true "请求信息"
// @Success 200 {object} response.GetAlertDetailResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/event/detail [post]
func (h *handler) AlertEventDetail() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertDetailRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.alertService.AlertDetail(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertEventListError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
