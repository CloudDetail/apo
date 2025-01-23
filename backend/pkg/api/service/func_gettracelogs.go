// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetTraceLogs 获取Trace故障现场日志
// @Summary 获取Trace故障现场日志
// @Description 获取Trace故障现场日志
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Param instance query string false "实例名"
// @Param nodeName query string false "主机名"
// @Param containerId query string false "容器名"
// @Param pid query uint32 false "进程号"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []clickhouse.FaultLogResult
// @Failure 400 {object} code.Failure
// @Router /api/service/trace/logs [get]
func (h *handler) GetTraceLogs() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTraceLogsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, 0, nil, &req.Service, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.HandleError(err, code.AuthError)
			return
		}
		resp, err := h.serviceInfoService.GetTraceLogs(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetTraceLogsError,
				code.Text(code.GetTraceLogsError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
