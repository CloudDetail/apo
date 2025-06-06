// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package license

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetCurrentLicense get current license.
// @Summary get current license.
// @Description get current license.
// @Tags API.license
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetCurrentLicenseResp
// @Failure 400 {object} code.Failure
// @Router /api/license/current [get]
func (h *handler) GetCurrentLicense() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.service.GetCurrentLicense()
		if err != nil {
			c.HandleError(err, code.GetLicenseError, response.GetCurrentLicenseResp{})
			return
		}

		c.Payload(resp)
	}
}
