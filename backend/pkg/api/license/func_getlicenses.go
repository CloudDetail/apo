// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package license

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetLicenses get all license.
// @Summary get all license.
// @Description get all license.
// @Tags API.license
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetLicenseResp
// @Failure 400 {object} code.Failure
// @Router /api/license [get]
func (h *handler) GetLicenses() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.service.GetLicenses()
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetLicenseError, err)
			return
		}

		c.Payload(resp)
	}
}
