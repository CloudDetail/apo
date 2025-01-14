// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"net/http"
)

func (m *middleware) PermissionMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		userID := GetContextUserID(c)
		if userID == 0 {
			c.AbortWithError(core.Error(http.StatusUnauthorized, code.UnAuth, code.Text(code.UnAuth)))
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
					code.Text(vErr.Code)).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusForbidden,
					code.AuthError,
					code.Text(code.AuthError)).WithError(err))
			}
			return
		}

		if !can {
			c.AbortWithError(core.Error(
				http.StatusForbidden,
				code.UserNoPermissionError,
				code.Text(code.UserNoPermissionError)))
			return
		}

		c.Next()
	}
}
