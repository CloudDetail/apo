// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// ListMetrics
// @Summary
// @Description
// @Tags API.metric
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/metric/list [get]
func (h *handler) ListMetrics() core.HandlerFunc {
	return func(c core.Context) {
		resp := h.metricService.ListPreDefinedMetrics()
		c.Payload(resp)
	}
}
