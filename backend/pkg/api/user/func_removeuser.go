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

// Remove user RemoveUser
// @Summary remove user
// @Description remove user
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Param userId formData int64 true "request information"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/remove [post]
func (h *handler) RemoveUser() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.RemoveUserRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.userService.RemoveUser(req.UserID)
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
					code.RemoveUserError,
					code.Text(code.RemoveUserError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
