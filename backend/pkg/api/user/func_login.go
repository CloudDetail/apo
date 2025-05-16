// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// Login
// @Summary Login
// @Description Login
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Success 200 {object} response.LoginResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/login [post]
func (h *handler) Login() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LoginRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.userService.Login(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UserLoginError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
