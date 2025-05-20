// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// RemoveUser Remove a user.
// @Summary Remove a user.
// @Description Remove a user.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Param userId formData int64 true "User's id"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/remove [post]
func (h *handler) RemoveUser() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.RemoveUserRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.userService.RemoveUser(req.UserID)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.RemoveUserError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
