// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// Logout Logout
// @Summary Logout
// @Description Logout
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param accessToken formData string true "accessToken"
// @Param refreshToken formData string true "refreshToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/logout [post]
func (h *handler) Logout() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogoutRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.userService.Logout(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.InValidToken,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
