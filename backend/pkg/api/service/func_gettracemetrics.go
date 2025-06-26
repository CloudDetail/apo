// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetTraceMetrics get Trace related metrics
// @Summary get Trace related metrics
// @Description get Trace related metrics
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
// @Success 200 {object} []response.GetTraceMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/trace/metrics [post]
func (h *handler) GetTraceMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTraceMetricsRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		userID := c.UserID()
		err := h.dataService.CheckDatasourcePermission(c, userID, 0, nil, &req.Service, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.AbortWithPermissionError(err, code.AuthError, []response.GetTraceMetricsResponse{})
			return
		}
		resp, err := h.serviceInfoService.GetTraceMetrics(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetTraceMetricsError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
