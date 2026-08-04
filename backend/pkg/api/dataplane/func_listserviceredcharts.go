// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// ListServiceRedCharts list services' redcharts
// @Summary list services' redcharts
// @Description list services' redcharts
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request query request.ListServiceRedChartsRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.ListServiceRedChartsResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/redcharts/list [get]
func (h *handler) ListServiceRedCharts() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ListServiceRedChartsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp := h.dataplaneService.ListServiceRedCharts(c, req)
		c.Payload(resp)
	}
}
