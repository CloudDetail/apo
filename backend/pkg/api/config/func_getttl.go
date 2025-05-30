// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetTTL Get TTL
// @Summary get TTL
// @Description get TTL
// @Tags API.config
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetTTLResponse
// @Failure 400 {object} code.Failure
// @Router /api/config/getTTL [get]
func (h *handler) GetTTL() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.configService.GetTTL(c)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetTTLError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
