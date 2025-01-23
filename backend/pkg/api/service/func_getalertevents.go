// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetAlertEvents 获取告警事件
// @Summary 获取告警事件
// @Description 获取告警事件
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string false "查询服务名"
// @Param source query string false "查询告警来源"
// @Param group query string false "查询告警类型"
// @Param name query string false "查询告警名称"
// @Param id query string false "查询告警ID"
// @Param status query string false "查询告警状态"
// @Param currentPage query int false "分页参数,当前页数"
// @Param pageSize query int false "分页参数, 每页数量"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetAlertEventsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/alert/events [get]
func (h *handler) GetAlertEvents() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertEventsRequest)
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
		resp, err := h.serviceInfoService.GetAlertEvents(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertEventsError,
				code.Text(code.GetAlertEventsError)).WithError(err),
			)
			return
		}
		if resp == nil {
			resp = &response.GetAlertEventsResponse{
				TotalCount: 0,
				EventList:  []clickhouse.PagedAlertEvent{},
			}
		}
		c.Payload(resp)
	}
}
