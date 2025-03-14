// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/services/metric"
)

// QueryMetrics
// @Summary
// @Description
// @Tags API.metric
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body metric.QueryMetricsRequest true "request"
// @Success 200 {object} metric.QueryMetricsResult
// @Failure 400 {object} code.Failure
// @Router /api/metric/query [post]
func (h *handler) QueryMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(metric.QueryMetricsRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp := h.metricService.QueryMetrics(req)
		c.Payload(resp)
	}
}
