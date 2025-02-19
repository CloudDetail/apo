// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetStaticIntegration
// @Summary
// @Description
// @Tags API.integration
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} response.getStoreIntegrationResponse
// @Failure 400 {object} code.Failure
// @Router /api/integration/configuration [get]
func (h *handler) GetStaticIntegration() core.HandlerFunc {
	return func(c core.Context) {
		storeIntegration := h.integrationService.GetStaticIntegration()
		c.Payload(storeIntegration)
	}
}
