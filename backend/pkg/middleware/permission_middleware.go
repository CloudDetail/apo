// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (m *middleware) PermissionMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		userID := GetContextUserID(c)
		if userID == 0 {
			c.AbortWithError(core.Error(http.StatusUnauthorized, code.UnAuth, c.ErrMessage(code.UnAuth)))
			return
		}
		method, path := c.GetMethodPath()

		can, err := m.permissionService.CheckApiPermission(userID, method, path)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusForbidden,
					vErr.Code,
					c.ErrMessage(vErr.Code)).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusForbidden,
					code.AuthError,
					c.ErrMessage(code.AuthError)).WithError(err))
			}
			return
		}

		if !can {
			c.AbortWithError(core.Error(
				http.StatusForbidden,
				code.UserNoPermissionError,
				c.ErrMessage(code.UserNoPermissionError)))
			return
		}

		c.Next()
	}
}
