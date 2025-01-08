// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"net/http"
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
					code.UserTokenExpireError,
					code.Text(code.UserTokenExpireError),
				).WithError(err))
			}
			return
		}
		c.Payload(resp)
	}
}
