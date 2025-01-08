// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// CreateUser Create a user.
// @Summary Create a user.
// @Description Create a user.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.CreateUserRequest true "Request"
// @Param Authorization header string false "Bearer token"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/create [post]
func (h *handler) CreateUser() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateUserRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if req.ConfirmPassword != req.Password {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserConfirmPasswordError,
				code.Text(code.UserConfirmPasswordError)),
			)
			return
		}

		if len(req.Phone) > 0 && !phoneRegexp.MatchString(req.Phone) {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserPhoneFormatError,
				code.Text(code.UserPhoneFormatError)))
			return
		}

		err := h.userService.CreateUser(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code),
				).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.UserCreateError,
					code.Text(code.UserCreateError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
