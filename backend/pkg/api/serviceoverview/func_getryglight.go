// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

// GetRYGLight 获取红绿灯结果
// @Summary 获取红绿灯结果
// @Description 获取红绿灯结果
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param serviceName query string false "服务名称"
// @Param endpointName query string false "接口名称"
// @Param namespace query string false "命名空间"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.ServiceRYGLightRes
// @Failure 400 {object} code.Failure
// @Router /api/service/ryglight [get]
func (h *handler) GetRYGLight() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetRygLightRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, &req.Namespace, &req.ServiceName)
		if err != nil {
			c.HandleError(err, code.AuthError)
			return
		}
		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)
		filter := serviceoverview.EndpointsFilter{
			ContainsSvcName:      req.ServiceName,
			ContainsEndpointName: req.EndpointName,
			Namespace:            req.Namespace,
		}

		resp, err := h.serviceoverview.GetServicesRYGLightStatus(startTime, endTime, filter)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceMoreUrlListError,
				code.Text(code.GetServiceMoreUrlListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
