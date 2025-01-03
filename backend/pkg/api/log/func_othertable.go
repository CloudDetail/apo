// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// OtherTable 获取外部日志表信息
// @Summary 获取外部日志表
// @Description 获取外部日志表
// @Tags API.log
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.OtherTableRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.OtherTableResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/other [get]
func (h *handler) OtherTable() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.OtherTableRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.OtherTable(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAllOtherLogTableError,
				code.Text(code.GetAllOtherLogTableError)+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
