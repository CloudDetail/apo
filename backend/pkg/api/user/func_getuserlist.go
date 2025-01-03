// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Param username query string false "用户名"
// @Param currentPage query string false "当前页"
// @Param pageSize query string false "页大小"
// @Param roleList query []int false "角色" collectionFormat(multi)
// @Param corporation query string false "组织"
// @Success 200 {object} response.GetUserListResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/list [get]
func (h *handler) GetUserList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetUserListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		if req.PageParam == nil {
			req.PageParam = &request.PageParam{
				CurrentPage: 1,
				PageSize:    10,
			}
		}

		resp, err := h.userService.GetUserList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetUserInfoError,
				code.Text(code.GetUserInfoError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
