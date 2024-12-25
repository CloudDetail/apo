// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetPolarisInfer 获取北极星指标分析情况
// @Summary 获取北极星指标分析情况
// @Description 获取北极星指标分析情况
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param step query int64 true "查询步长(us)"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetPolarisInferResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/polaris/infer [get]
func (h *handler) GetPolarisInfer() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetPolarisInferRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		res, err := h.serviceInfoService.GetPolarisInfer(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetPolarisInferError,
				code.Text(code.GetPolarisInferError)).WithError(err),
			)
			return
		}

		c.Payload(res)
	}
}
