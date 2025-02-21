// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

const (
	UserIDKey = "userID"
)

func (m *middleware) AuthMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		rawToken := c.GetHeader("Authorization")
		token := util.ParseRawToken(rawToken)

		if len(token) == 0 {
			if !config.Get().User.AnonymousUser.Enable {
				c.AbortWithError(core.Error(http.StatusBadRequest, code.UnAuth, code.Text(code.UnAuth)))
				return
			}

			anonymousUser, err := m.userService.GetUserInfo(0)
			if err != nil {
				c.AbortWithError(core.Error(http.StatusBadRequest, code.AuthError, code.Text(code.AuthError)))
				return
			}

			c.Set(UserIDKey, anonymousUser.UserID)
			c.Next()
			return
		}

		if ok, _ := m.userService.IsInBlacklist(token); ok {
			c.AbortWithError(core.Error(http.StatusBadRequest, code.InValidToken, code.Text(code.InValidToken)))
			return
		}

		claims, err := util.ParseAccessToken(token)
		if err != nil {
			c.AbortWithError(core.Error(http.StatusBadRequest, code.InValidToken, code.Text(code.InValidToken)))
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

func GetContextUserID(c core.Context) int64 {
	userID, ok := c.Get(UserIDKey)
	if !ok {
		return 0
	}
	id, ok := userID.(int64)
	if !ok {
		return 0
	}
	return id
}
