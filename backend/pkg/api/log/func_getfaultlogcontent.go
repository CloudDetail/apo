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

// GetFaultLogContent 获取故障现场日志内容
// @Summary 获取故障现场日志内容
// @Description 获取故障现场日志内容
// @Tags API.log
// @Produce json
// @Param Request body request.GetFaultLogContentRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetFaultLogContentResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/fault/content [post]
func (h *handler) GetFaultLogContent() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetFaultLogContentRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, 0, nil, &req.ServiceName, "")
		if err != nil {
			c.HandleError(err, code.AuthError)
			return
		}
		resp, err := h.logService.GetFaultLogContent(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFaultLogContentError,
				code.Text(code.GetFaultLogContentError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
