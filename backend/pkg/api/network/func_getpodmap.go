// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package deepflow

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"
)

// GetPodMap 查询 Pod 网络调用拓扑与调用指标
// @Summary 查询 Pod 网络调用拓扑与调用指标
// @Description 查询 Pod 网络调用拓扑与调用指标
// @Tags API.Network
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "起始时间，单位微秒"
// @Param endTime query int64 true "结束时间，单位微秒"
// @Param namespace query string false "要查询的Namespace，值为空则查询所有"
// @Param workload query string false "要查询的工作负载，值为空则查询所有"
// @Success 200 {object} response.PodMapResponse
// @Failure 400 {object} code.Failure
// @Router /api/network/podmap [get]
func (h *handler) GetPodMap() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.PodMapRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, &req.Namespace, nil)
		if err != nil {
			c.HandleError(err, code.AuthError)
			return
		}
		resp, err := h.networkService.GetPodMap(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ServerError,
				code.Text(code.ServerError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
