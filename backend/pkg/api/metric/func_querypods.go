// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package metric

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryPods queries pods observed in metrics.
// @Summary Query pods from metrics
// @Description Queries pods observed in the specified time range and optionally filters them by node name, namespace, or pod name.
// @Tags API.metric
// @Accept json
// @Produce json
// @Param Request body request.QueryPodsRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.QueryPodsResponse
// @Failure 400 {object} code.Failure
// @Router /api/metric/queryPods [post]
func (h *handler) QueryPods() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.QueryPodsRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.metricService.QueryPods(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.QueryPodsError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
