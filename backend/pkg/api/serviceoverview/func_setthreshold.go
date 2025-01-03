// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// SetThreshold @Summary 配置单个阈值配置信息
// @Summary 配置单个阈值配置信息
// @Description 配置单个阈值配置信息
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param level formData string true "阈值等级"
// @Param serviceName formData string false "应用名称"
// @Param endpoint formData string false "endpoint"
// @Param latency formData float64 true "同比延时"
// @Param errorRate formData float64 true "同比错误率"
// @Param tps formData float64 true "同比请求次数"
// @Param log formData float64 true "同比日志告警"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200
// @Failure 400 {object} code.Failure
// @Router /api/service/setThreshold [post]
func (h *handler) SetThreshold() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetThresholdRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		level := req.Level
		serviceName := req.ServiceName
		endpoint := req.Endpoint
		latency := req.Latency
		errorRate := req.ErrorRate
		tps := req.Tps
		log := req.Log
		resp, err := h.serviceoverview.SetThreshold(level, serviceName, endpoint, latency, errorRate, tps, log)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.SetThresholdError,
				code.Text(code.SetThresholdError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
