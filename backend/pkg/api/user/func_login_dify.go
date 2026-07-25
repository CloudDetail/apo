// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// LoginDify re-establishes Dify credentials using the authenticated APO user.
// @Summary Login to Dify with the current APO session
// @Tags API.user
// @Produce json
// @Success 200 {object} dify.DifyLoginResponse
// @Failure 502 {object} code.Failure
// @Router /api/user/dify/login [post]
func (h *handler) LoginDify() core.HandlerFunc {
	return func(c core.Context) {
		if c.UserID() == 0 {
			c.AbortWithError(http.StatusBadRequest, code.UnAuth, nil)
			return
		}
		resp, err := h.userService.LoginDify(c, c.UserID())
		if err != nil {
			c.AbortWithError(http.StatusBadGateway, code.UserLoginError, err)
			return
		}
		c.Payload(resp)
	}
}
