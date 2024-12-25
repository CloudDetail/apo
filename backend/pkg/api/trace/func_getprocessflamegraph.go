// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetProcessFlameGraph 获取并整合进程级别火焰图数据
// @Summary 获取并整合进程级别火焰图数据
// @Description 获取并整合进程级别火焰图数据
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param maxNodes query int64 false "限制节点数"
// @Param startTime query int64 true "开始时间"
// @Param endTime query int64 true "结束时间"
// @Param pid query int64 true "进程id"
// @Param nodeName query string false "主机名称"
// @Param sampleType query string true "采样类型"
// @Success 200 {object} response.GetProcessFlameGraphResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/flame/process [get]
func (h *handler) GetProcessFlameGraph() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetProcessFlameGraphRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.traceService.GetProcessFlameGraphData(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFlameGraphError,
				code.Text(code.GetFlameGraphError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
