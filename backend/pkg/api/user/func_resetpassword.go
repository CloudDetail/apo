// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"
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
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		if req.NewPassword != req.ConfirmPassword {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserConfirmPasswdError,
				c.ErrMessage(code.UserConfirmPasswdError)),
			)
			return
		}

		err := h.userService.RestPassword(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					c.ErrMessage(vErr.Code),
				).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.UserUpdateError,
					c.ErrMessage(code.UserUpdateError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
