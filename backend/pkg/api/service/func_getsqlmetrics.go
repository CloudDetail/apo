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

// Get SQL metrics GetSQLMetrics
// @Summary get SQL metrics
// @Description get SQL metrics
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param step query int64 true "query step (us)"
// @Param sortBy query string true "Sort method (latency/errorRate/tps)"
// @Param currentPage query int false "Paging parameter, current number of pages"
// @Param pageSize query int false "Pagination parameter, quantity per page"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetSQLMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/sql/metrics [post]
func (h *handler) GetSQLMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSQLMetricsRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allow, err := h.dataService.CheckGroupPermission(c, req.GroupID); !allow || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, &response.GetSQLMetricsResponse{
				Pagination: model.Pagination{
					Total:       0,
					CurrentPage: 0,
					PageSize:    0,
				},
				SQLOperationDetails: []response.SQLOperationDetail{},
			})
			return
		}

		resp, err := h.serviceInfoService.GetSQLMetrics(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetSQLMetricError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
