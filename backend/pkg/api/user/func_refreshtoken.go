// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// RefreshToken Refresh accessToken
// @Summary Refresh accessToken
// @Description Refresh accessToken
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "Bearer refreshToken"
// @Success 200 {object} response.RefreshTokenResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/refresh [get]
func (h *handler) RefreshToken() core.HandlerFunc {
	return func(c core.Context) {
		token := c.GetHeader("Authorization")

		resp, err := h.userService.RefreshToken(token)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UserTokenExpireError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
