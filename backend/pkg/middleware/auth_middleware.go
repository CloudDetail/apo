// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"net/http"
)

const (
	UserIDKey = "userID"
)

func (m *middleware) AuthMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		rawToken := c.GetHeader("Authorization")
		token := util.ParseRawToken(rawToken)
		if len(token) == 0 {
			if config.Get().User.AnonymousUser.Enable {
				c.Next()
				return
			} else {
				c.AbortWithError(core.Error(http.StatusBadRequest, code.UnAuth, code.Text(code.UnAuth)))
				return
			}
		}

		// TODO handle error when switch to redis
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
