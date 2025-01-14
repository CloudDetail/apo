// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetServiceInstance 获取service对应url实例
// @Summary 获取service对应url实例
// @Description 获取service对应url实例
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param step query int64 true "步长"
// @Param serviceName query string true "应用名称"
// @Param endpoint query string false "endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.InstancesRes
// @Failure 400 {object} code.Failure
// @Router /api/service/instances [get]
func (h *handler) GetServiceInstance() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceInstanceRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, nil, &req.ServiceName)
		if err != nil {
			c.HandleError(err, code.AuthError)
			return
		}
		var startTime time.Time
		var endTime time.Time
		req.StartTime = req.StartTime / 1000000 //接收的微秒级别的startTime和endTime需要先转成秒级别
		req.EndTime = req.EndTime / 1000000     //接收的微秒级别的startTime和endTime需要先转成秒级别
		startTime = time.Unix(req.StartTime, 0)
		endTime = time.Unix(req.EndTime, 0)
		step := time.Duration(req.Step * 1000)
		serviceName := req.ServiceName
		endpoint := req.Endpoint
		data, err := h.serviceInfoService.GetInstancesNew(startTime, endTime, step, serviceName, endpoint)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetOverviewServiceInstanceListError,
				code.Text(code.GetOverviewServiceInstanceListError)).WithError(err),
			)
			return
		}
		c.Payload(data)
	}
}
