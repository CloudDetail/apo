// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// UpdateUserInfo Update user's info.
// @Summary Update user's info.
// @Description Update user's info.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 true "User's id"
// @Param roleList formData []int false "The list of user's role." collectionFormat(multi)
// @Param corporation formData string false "Corporation"
// @Param phone formData string false "Phone number"
// @Param email formData string false "Email"
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/update/info [post]
func (h *handler) UpdateUserInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateUserInfoRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.userService.UpdateUserInfo(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UserUpdateError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
