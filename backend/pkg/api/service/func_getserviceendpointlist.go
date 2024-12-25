// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServiceEndPointList 获取服务EndPoint列表
// @Summary 获取服务EndPoint列表
// @Description 获取服务EndPoint列表
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string false "查询服务名"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []string
// @Failure 400 {object} code.Failure
// @Router /api/service/endpoint/list [get]
func (h *handler) GetServiceEndPointList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceEndPointListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.GetServiceEndPointList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceEndPointListError,
				code.Text(code.GetServiceEndPointListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
