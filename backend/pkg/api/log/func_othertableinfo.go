// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// OtherTableInfo 获取外部日志表信息
// @Summary 获取外部日志表信息
// @Description 获取外部日志表信息
// @Tags API.log
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.OtherTableInfoRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.OtherTableInfoResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/other/table [get]
func (h *handler) OtherTableInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.OtherTableInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.OtherTableInfo(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetOtherLogTableError,
				code.Text(code.GetOtherLogTableError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
