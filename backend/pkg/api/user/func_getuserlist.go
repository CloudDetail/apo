// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetUserList get user list
// @Summary get user list
// @Description get user list
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Param username query string false "username"
// @Param currentPage query string false "current page"
// @Param pageSize query string false "Page size"
// @Param role query string false "role"
// @Param corporation query string false "organization"
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
				PageSize:    99,
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
