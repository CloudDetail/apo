// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// SetThreshold @Summary configuration single threshold configuration information
// @Summary configuration single threshold configuration information
// @Description configuration single threshold configuration information
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param level formData string true "threshold level"
// @Param serviceName formData string false "app name"
// @Param endpoint formData string false "endpoint"
// @Param latency formData float64 true "YoY Delay"
// @Param errorRate formData float64 true "YoY Error Rate"
// @Param tps formData float64 true "Number of requests compared with the same period last year"
// @Param log formData float64 true "year-on-year log alarm"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200
// @Failure 400 {object} code.Failure
// @Router /api/service/setThreshold [post]
func (h *handler) SetThreshold() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetThresholdRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
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
		resp, err := h.serviceoverview.SetThreshold(c, level, serviceName, endpoint, latency, errorRate, tps, log)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.SetThresholdError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
