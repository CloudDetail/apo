// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// UpdateSelfInfo Update self info.
// @Summary Update self info.
// @Description Update self info.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 true "User's id"
// @Param corporation formData string false "Corporation"
// @Param phone formData string false "Phone number"
// @Param email formData string false "Email"
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/update/self [post]
func (h *handler) UpdateSelfInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateSelfInfoRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.userService.UpdateSelfInfo(c, req)
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
