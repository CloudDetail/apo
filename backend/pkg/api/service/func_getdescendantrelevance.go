// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetDescendantRelevance 获取依赖节点延时关联度
// @Summary 获取依赖节点延时关联度
// @Description 获取依赖节点延时关联度
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Param step query int64 true "查询步长(us)"
// @Param entryService query string false "入口服务名"
// @Param entryEndpoint query string false "入口Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []response.GetDescendantRelevanceResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/descendant/relevance [get]
func (h *handler) GetDescendantRelevance() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDescendantRelevanceRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, nil, &req.Service)
		if err != nil {
			c.HandleError(err, code.AuthError)
			return
		}
		res, err := h.serviceInfoService.GetDescendantRelevance(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetDescendantRelevanceError,
				code.Text(code.GetDescendantRelevanceError)).WithError(err),
			)
			return
		}
		c.Payload(res)
	}
}
