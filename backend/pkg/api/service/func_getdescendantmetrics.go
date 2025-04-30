// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetDescendantMetrics get the delay curve data of all downstream services
// @Summary get the delay curve data of all downstream services
// @Description get the delay curve data of all downstream services
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param endpoint query string true "Query Endpoint"
// @Param step query int64 true "query step (us)"
// @Param entryService query string false "Ingress service name"
// @Param entryEndpoint query string false "entry Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []response.GetDescendantMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/descendant/metrics [get]
func (h *handler) GetDescendantMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDescendantMetricsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, 0, nil, &req.Service, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.HandleError(err, code.AuthError, nil)
			return
		}
		resp, err := h.serviceInfoService.GetDescendantMetrics(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetDescendantMetricsError,
				c.ErrMessage(code.GetDescendantMetricsError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
