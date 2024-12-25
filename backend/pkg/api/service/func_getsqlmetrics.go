// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetSQLMetrics 获取SQL指标
// @Summary 获取SQL指标
// @Description 获取SQL指标
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param step query int64 true "查询步长(us)"
// @Param sortBy query string true "排序方法(latency/errorRate/tps)"
// @Param currentPage query int false "分页参数,当前页数"
// @Param pageSize query int false "分页参数, 每页数量"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetSQLMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/sql/metrics [get]
func (h *handler) GetSQLMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSQLMetricsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.GetSQLMetrics(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetSQLMetricError,
				code.Text(code.GetSQLMetricError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
