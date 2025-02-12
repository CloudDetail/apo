// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

func (m *middleware) LicMiddleware() core.HandlerFunc {
	return func(c core.Context) {
		conf := config.Get()
		if conf.License.BasicLicense == nil {
			err := conf.ParseLicense()
			if err != nil {
				c.AbortWithError(core.Error(http.StatusInternalServerError, code.GetLicenseError, code.Text(code.GetLicenseError)).WithError(err))
				return
			}
		}
		var requiredFeatures []string
		// TODO get api
		// TODO api feature api
		// feature, err := m.PermissionService.GetFeatureByAPI(c.GetAPI())
		err := conf.License.BasicLicense.CheckFeatures(requiredFeatures...)
		if err != nil {
			c.AbortWithError(core.Error(http.StatusInternalServerError, code.LicenseFeatureValidationError, code.Text(code.LicenseFeatureValidationError)))
		}
		c.Next()
	}
}
