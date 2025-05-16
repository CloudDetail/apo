// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetUserList Get user list.
// @Summary Get user list.
// @Description Get user list.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Param username query string false "Username"
// @Param currentPage query string false "Current page"
// @Param pageSize query string false "The size of page"
// @Param roleList query []int false "Role id list" collectionFormat(multi)
// @Param teamList query []int false "Team id list" collectionFormat(multi)
// @Param corporation query string false "组织"
// @Success 200 {object} response.GetUserListResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/list [get]
func (h *handler) GetUserList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetUserListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
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
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetUserInfoError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
