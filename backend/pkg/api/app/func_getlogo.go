// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package app

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetLogo gets logo
// @Summary gets logo
// @Description gets logo
// @Tags API.app
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetLogoResp
// @Failure 400 {object} code.Failure
// @Router /api/app/logo [get]
func (h *handler) GetLogo() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.service.GetLogo()
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetLogoError,
				err,
			)
		}
		c.Payload(resp)
	}
}
