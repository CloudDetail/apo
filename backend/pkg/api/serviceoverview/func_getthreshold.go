// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetThreshold get the configuration information of a single threshold
// @Summary get individual threshold configuration information
// @Description get individual threshold configuration information
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param serviceName query string false "app name"
// @Param endpoint query string false "endpoint"
// @Param level query string true "Threshold level"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200
// @Failure 400 {object} code.Failure
// @Router /api/service/getThreshold [get]
func (h *handler) GetThreshold() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetThresholdRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}
		serviceName := req.ServiceName
		endPoint := req.Endpoint
		level := req.Level
		resp, err := h.serviceoverview.GetThreshold(level, serviceName, endPoint)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetThresholdError,
				c.ErrMessage(code.GetThresholdError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
