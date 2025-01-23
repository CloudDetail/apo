// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServicesAlert 获取Service的日志告警和状态灯信息
// @Summary 获取Service的日志告警和状态灯信息
// @Description 获取Service的日志告警和状态灯信息
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param step query int64 true "步长"
// @Param serviceNames query []string true "应用名称" collectionFormat(multi)
// @Param returnData query []string false "返回数据内容" collectionFormat(multi)
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.ServiceAlertRes
// @Failure 400 {object} code.Failure
// @Router /api/service/servicesAlert [get]
func (h *handler) GetServicesAlert() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceAlertRequest)

		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, 0, nil, &req.ServiceNames, model.DATASOURCE_CATEGORY_APM)
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
		serviceNames := req.ServiceNames
		returnData := req.ReturnData
		var resp []response.ServiceAlertRes
		data, err := h.serviceoverview.GetServicesAlert(startTime, endTime, step, serviceNames, returnData)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServicesAlertError,
				code.Text(code.GetServicesAlertError)).WithError(err),
			)
			return
		}
		if data != nil {
			resp = data
		} else {
			resp = []response.ServiceAlertRes{}
		}

		c.Payload(resp)
	}
}
