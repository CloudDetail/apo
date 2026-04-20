// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package app

import (
	"os"

	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetDeployMode Get deploy mode.
// @Summary Get deploy mode.
// @Description Get deploy mode.
// @Tags API.app
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/app/deployMode [get]
func (h *handler) GetDeployMode() core.HandlerFunc {
	return func(c core.Context) {
		deployMode := os.Getenv("PRODUCT_MODE")
		c.Payload(deployMode)
	}
}
