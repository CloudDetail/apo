// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetAlertEventStaticFilters
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} response.AlertEventFiltersResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/filter/keys [get]
func (h *handler) GetAlertEventStaticFilters() core.HandlerFunc {
	return func(c core.Context) {
		resp := h.alertService.GetStaticFilterKeys(c)
		c.Payload(resp)
	}
}
