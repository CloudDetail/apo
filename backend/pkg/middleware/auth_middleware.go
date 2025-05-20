// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/util/jwt"
)

func (m *middleware) AuthMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		rawToken := c.GetHeader("Authorization")
		token := jwt.ParseRawToken(rawToken)

		if len(token) == 0 {
			if !config.Get().User.AnonymousUser.Enable {
				c.AbortWithError(http.StatusBadRequest, code.UnAuth, nil)
				return
			}

			anonymousUser, err := m.userService.GetUserInfo(c, 0)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, code.AuthError, nil)
				return
			}

			c.Set(core.UserIDKey, anonymousUser.UserID)
			c.Next()
			return
		}

		if ok, _ := m.userService.IsInBlacklist(c, token); ok {
			c.AbortWithError(http.StatusBadRequest, code.InValidToken, nil)
			return
		}

		claims, err := jwt.ParseAccessToken(token)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.InValidToken, nil)
			return
		}

		c.Set(core.UserIDKey, claims.UserID)
		c.Next()
	}
}
