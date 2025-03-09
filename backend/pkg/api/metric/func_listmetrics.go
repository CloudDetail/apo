// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// TODO 迁移到 model/request包中
type listMetricsRequest struct {
}

// TODO 迁移到 model/response包中
type listMetricsResponse struct {
}

// ListMetrics
// @Summary
// @Description
// @Tags API.metric
// @Accept application/x-www-form-urlencoded
// @Produce json
// TODO 下面的请求参数类型和返回类型需根据实际需求进行变更
// @Param Request body request.listMetricsRequest true "请求信息"
// @Success 200 {object} response.listMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/metric/list [get]
func (h *handler) ListMetrics() core.HandlerFunc {
	return func(c core.Context) {
		// TODO 替换为Service调用
		resp := h.metricService.ListPreDefinedMetrics()
		c.Payload(resp)
	}
}
