// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServiceRoute 获取服务对应的应用日志
// @Summary 获取服务对应的应用日志
// @Description 获取服务对应的应用日志
// @Tags API.log
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetServiceRouteRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetServiceRouteResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/rule/service [get]
func (h *handler) GetServiceRoute() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceRouteRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, 0, nil, &req.Service, "")
		if err != nil {
			c.HandleError(err, code.AuthError)
			return
		}
		resp, err := h.logService.GetServiceRoute(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceRouteError,
				code.Text(code.GetServiceRouteError)+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
