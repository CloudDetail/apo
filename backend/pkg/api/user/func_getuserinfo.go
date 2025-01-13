// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/middleware"
)

// GetUserInfo access to personal information
// @Summary get personal information
// @Description get personal information
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetUserInfoResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/info [get]
func (h *handler) GetUserInfo() core.HandlerFunc {
	return func(c core.Context) {

		userID := middleware.GetContextUserID(c)
		resp, err := h.userService.GetUserInfo(userID)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetUserInfoError,
				code.Text(code.GetUserInfoError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
