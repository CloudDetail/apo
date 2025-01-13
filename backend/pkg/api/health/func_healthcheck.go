// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// HealthCheck for k8s to check backend health status
// @Summary for k8s to check backend health status
// @Description for k8s to check backend health status
// @Tag API.health
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/health [get]
func (h *handler) HealthCheck() core.HandlerFunc {
	return func(c core.Context) {
		c.Payload("ok")
	}
}
