// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServiceRoute get the application log corresponding to the service
// @Summary get the application log corresponding to the service
// @Description get the application log corresponding to the service
// @Tags API.log
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetServiceRouteRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetServiceRouteResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/service [get]
func (h *handler) GetServiceRoute() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceRouteRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.logService.GetServiceRoute(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetServiceRouteError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
