// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// List xx列表
// @Summary xx列表
// @Description xx列表
// @Tags API.mock
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param page_num query int true "第几页" default(1)
// @Param page_size query int false "每页显示条数" default(10)
// @Param name query string false "用户名"
// @Success 200 {object} response.ListResponse
// @Failure 400 {object} code.Failure
// @Router /api/mock [get]
func (h *handler) List() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		if req.PageSize == 0 {
			req.PageSize = 10
		}

		resp, err := h.mockService.PageList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.MockListError,
				code.Text(code.MockListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
