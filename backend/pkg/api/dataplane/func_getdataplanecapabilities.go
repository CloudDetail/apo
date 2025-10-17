// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetDataplaneCapabilities Get intergrated dataplane capabilities.
// @Summary Get intergrated dataplane capabilities.
// @Description Get intergrated dataplane capabilities.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetIntegratedCapabilitiesResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/capabilities [get]
func (h *handler) GetIntegratedCapabilities() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.dataplaneService.GetIntegratedCapabilities(c)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, code.GetIntegratedCapabilitiesError, err)
			return
		}
		c.Payload(resp)
	}
}
