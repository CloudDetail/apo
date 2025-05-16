// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// ResetPassword Reset user's password.
// @Summary Reset user's password.
// @Description Reset user's password.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 true "User's id"
// @Param newPassword formData string true "New password"
// @Param confirmPassword formData string true "Confirm password"
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/reset [post]
func (h *handler) ResetPassword() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ResetPasswordRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if req.NewPassword != req.ConfirmPassword {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UserConfirmPasswdError,
				nil,
			)
			return
		}

		err := h.userService.RestPassword(req)
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
