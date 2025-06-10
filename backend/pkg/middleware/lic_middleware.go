// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/license"
	licenseservice "github.com/CloudDetail/apo/backend/pkg/services/license"
)

var apiFeatureBlackList = map[string]string{
	"/api/license":         "license管理",
	"/api/license/current": "license管理",
	"/api/license/upload":  "license管理",
}

// LicMiddleware 访问对应功能时检查license是否过期以及是否可以访问对应功能
func (m *middleware) LicMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		if !license.ActivateLicenseCheck {
			c.Next()
			return
		}

		_, path := c.GetMethodPath()
		expire, err := m.licenseService.CheckLicenseDate()
		if err != nil {
			if !errors.Is(err, licenseservice.ErrNoLicense) {
				c.AbortWithError(
					http.StatusBadRequest, code.LicenseCheckError, err,
				)
				return
			} else {
				if path == "/api/license/upload" {
					c.Next()
					return
				} else {
					c.AbortWithError(
						http.StatusInternalServerError, code.LicenseCheckError, err,
					)
				}
			}
		}

		if !expire {
			has, err := m.checkFeature(path)
			if err != nil {
				c.AbortWithError(
					http.StatusInternalServerError, code.LicenseCheckError, err,
				)
				return
			}

			if !has {
				c.AbortWithError(
					http.StatusInternalServerError, code.LicenseFeatureValidationError, err,
				)
				return
			}
			c.Next()
			return
		}

		activation, err := m.licenseService.GetActivation()
		if activation.Mode == "private" {
			c.AbortWithError(
				http.StatusBadRequest, code.LicenseExpired, err,
			)
			return
		}

		has, err := m.checkFeature(path)
		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError, code.LicenseCheckError, err,
			)
			return
		}

		if !has {
			c.AbortWithError(
				http.StatusInternalServerError, code.LicenseFeatureValidationError, err,
			)
			return
		}
		c.Next()
		return
	}
}

func (m *middleware) checkFeature(path string) (has bool, err error) {
	feature, ok := apiFeatureBlackList[path]
	if !ok {
		return true, nil
	}

	return m.licenseService.CheckFeature(feature)
}
