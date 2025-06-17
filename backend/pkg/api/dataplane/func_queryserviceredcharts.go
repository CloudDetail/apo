// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryServiceRedCharts Get service's redcharts.
// @Summary Get service's redcharts.
// @Description Get service's redcharts.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param cluster query string false "Query cluster name"
// @Param endpoint query string false "Query endpoint"
// @Success 200 {object} response.QueryServiceRedChartsResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/redcharts [get]
func (h *handler) QueryServiceRedCharts() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.QueryServiceRedChartsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if req.Endpoint == "" {
			c.Payload(h.dataplaneService.GetServiceRedCharts(c, req))
		} else {
			c.Payload(h.dataplaneService.GetServiceEndpointRedCharts(c, req))
		}
	}
}
