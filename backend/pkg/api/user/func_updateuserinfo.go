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

// UpdateUserInfo update personal information
// @Summary update personal information
// @Description update personal information
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 true "user id"
// @Param corporation formData string false "organization"
// @Param phone formData string false "Phone number"
// @Param email formData string false "mailbox"
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/update/info [post]
func (h *handler) UpdateUserInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateUserInfoRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.userService.UpdateUserInfo(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code)).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.UserUpdateError,
					code.Text(code.UserUpdateError)).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
