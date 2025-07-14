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
		authMethods := []authBy{
			authByToken,
			authByAnonymousUser,
		}

		for _, auth := range authMethods {
			ok, err := auth(c, m)
			if ok {
				return // auth success
			}

			if err != nil {
				c.AbortWithError(
					http.StatusBadRequest,
					code.UnAuth,
					err,
				)
				return // auth failed with error
			}
		}

		// unauth
		c.AbortWithError(
			http.StatusBadRequest,
			code.UnAuth,
			nil,
		)
	}
}

type authBy func(c core.Context, m *middleware) (bool, error)

func authByToken(c core.Context, m *middleware) (ok bool, err error) {
	rawToken := c.GetHeader("Authorization")
	token := jwt.ParseRawToken(rawToken)

	if len(token) == 0 {
		return false, nil
	}

	if ok, _ := m.userService.IsInBlacklist(c, token); ok {
		return false, core.Error(code.InValidToken, "")
	}

	claims, err2 := jwt.ParseAccessToken(token)
	if err2 != nil {
		return false, core.Error(code.InValidToken, "")
	}

	c.Set(core.UserIDKey, claims.UserID)
	c.Next()
	return true, nil
}

func authByAnonymousUser(c core.Context, m *middleware) (bool, error) {
	if !config.Get().User.AnonymousUser.Enable {
		return false, nil
	}

	anonymousUser, err := m.userService.GetUserInfo(c, 0)
	if err != nil {
		return false, core.Error(code.AuthError, "").WithStack(err)
	}
	c.Set(core.UserIDKey, anonymousUser.UserID)
	c.Next()
	return true, nil
}

func GetContextUserID(c core.Context) int64 {
	userID, ok := c.Get(core.UserIDKey)
	if !ok {
		return 0
	}
	id, ok := userID.(int64)
	if !ok {
		return 0
	}
	return id
}
