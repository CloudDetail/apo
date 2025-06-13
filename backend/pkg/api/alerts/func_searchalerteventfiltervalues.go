// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// SearchAlertEventFilterValues
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/json
// @Produce json
// @Param Request body request.SearchAlertEventFilterValuesRequest true "请求信息"
// @Success 200 {object} request.AlertEventFilter
// @Failure 400 {object} code.Failure
// @Router /api/alerts/filter/values [post]
func (h *handler) SearchAlertEventFilterValues() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SearchAlertEventFilterValuesRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.alertService.GetAlertEventFilterValues(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.SearchAlertFilterValueError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
