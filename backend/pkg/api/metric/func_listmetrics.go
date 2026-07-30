// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// ListMetrics lists the predefined metrics available for query.
// @Summary List predefined metrics
// @Description Returns the predefined metric catalog, including metric IDs, names, descriptions, required parameters, and units.
// @Tags API.metric
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []QueryInfo
// @Failure 400 {object} code.Failure
// @Router /api/metric/list [get]
func (h *handler) ListMetrics() core.HandlerFunc {
	return func(c core.Context) {
		resp := h.metricService.ListPreDefinedMetrics(c)
		c.Payload(resp)
	}
}
