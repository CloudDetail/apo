// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetLogMetrics get log metrics
// @Summary get log related metrics
// @Description get log related metrics
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query uint64 true "query start time"
// @Param endTime query uint64 true "query end time"
// @Param service query string true "Query service name"
// @Param endpoint query string true "Query Endpoint"
// @Param step query int64 true "query step (us)"
// @Param entryService query string false "Ingress service name"
// @Param entryEndpoint query string false "entry Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []response.GetLogMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/log/metrics [get]
func (h *handler) GetLogMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetLogMetricsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.GetLogMetrics(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetLogMetricsError,
				code.Text(code.GetLogMetricsError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
