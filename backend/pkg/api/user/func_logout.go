// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
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
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.userService.Logout(req)
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
					code.InValidToken,
					code.Text(code.InValidToken),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
