// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package deepflow

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// Segmentation metric of the time consumed by the GetSpanSegmentsMetrics client to call the Span network.
// @Segment metric of the time consumed by the Summary client to call the Span network
// @Segment metric of the time consumed by the Description client to call the Span network
// @Tags API.Network
// @Accept application/x-www-form-urlencoded
// @Param traceId query string true "traceId"
// @Param spanId query string false "spanId. If the value is blank, all items are queried"
// @Success 200 {object} response.SpanSegmentMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/network/segments [get]
func (h *handler) GetSpanSegmentsMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SpanSegmentMetricsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp, err := h.networkService.GetSpanSegmentsMetrics(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
