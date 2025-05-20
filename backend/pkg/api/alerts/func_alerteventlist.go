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
// @Param Request body request.AlertEventSearchRequest true "请求信息"
// @Success 200 {object} response.AlertEventSearchResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/events/list [post]
func (h *handler) AlertEventList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AlertEventSearchRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		list, err := h.alertService.AlertEventList(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertEventListError,
				err,
			)
			return
		}
		c.Payload(list)
	}
}
