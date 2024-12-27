// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DEPRECATED
// GetServiceInstanceList 获取服务实例列表
// @Summary 获取服务实例列表
// @Description 获取服务实例列表
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []string
// @Failure 400 {object} code.Failure
// @Deprecated
// @Router /api/service/instance/list [get]
func (h *handler) GetServiceInstanceList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceInstanceListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.GetServiceInstanceList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceInstanceListError,
				code.Text(code.GetServiceInstanceListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
