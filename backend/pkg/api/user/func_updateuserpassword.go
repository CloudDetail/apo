// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UpdateUserPassword Update password.
// @Summary Update password.
// @Description Update password.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 true "User's id"
// @Param oldPassword formData string true "Original password"
// @Param newPassword formData string true "New password"
// @Param confirmPassword formData string true "Confirm password"
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/update/password [post]
func (h *handler) UpdateUserPassword() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateUserPasswordRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if req.ConfirmPassword != req.NewPassword {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UserConfirmPasswdError,
				nil,
			)
			return
		}

		err := h.userService.UpdateUserPassword(c, req)
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
