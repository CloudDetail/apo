// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetFaultLogPageList 获取故障现场分页日志
// @Summary 获取故障现场分页日志
// @Description 获取故障现场分页日志
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.GetFaultLogPageListRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetFaultLogPageListResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/fault/pagelist [post]
func (h *handler) GetFaultLogPageList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetFaultLogPageListRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if req.PageNum == 0 {
			req.PageNum = 1
		}
		if req.PageSize == 0 {
			req.PageSize = 10
		}
		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, &req.Namespaces, &req.Service)
		if err != nil {
			c.HandleError(err, code.AuthError)
			return
		}
		resp, err := h.logService.GetFaultLogPageList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFaultLogPageListError,
				code.Text(code.GetFaultLogPageListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
