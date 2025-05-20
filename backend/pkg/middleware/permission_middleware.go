// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

func (m *middleware) PermissionMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		userID := c.UserID()
		if userID == 0 {
			c.AbortWithError(http.StatusUnauthorized, code.UnAuth, nil)
			return
		}
		method, path := c.GetMethodPath()

		can, err := m.permissionService.CheckApiPermission(c, userID, method, path)
		if err != nil {
			c.AbortWithError(http.StatusForbidden, code.AuthError, err)
			return
		}

		if !can {
			c.AbortWithError(http.StatusForbidden, code.UserNoPermissionError, nil)
			return
		}

		c.Next()
	}
}
