// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetOnOffCPU 获取span执行消耗
// @Summary 获取span执行消耗
// @Description 获取span执行消耗
// @Tags API.trace
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "开始时间"
// @Param endTime query int64  true "结束时间"
// @Param pid query uint32 true "进程id"
// @Param nodeName query string true "节点名"
// @Success 200 {object} response.GetOnOffCPUResponse
// @Failure 400 {object} code.Failure
// @Router /api/trace/onoffcpu [get]
func (h *handler) GetOnOffCPU() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetOnOffCPURequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.traceService.GetOnOffCPU(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetOnOffCPUError,
				code.Text(code.GetOnOffCPUError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
