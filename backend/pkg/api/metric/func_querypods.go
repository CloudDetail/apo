// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package metric

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryPods
// @Summary
// @Description
// @Tags API.metric
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.QueryPodsRequest true "Request information"
// @Produce json
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
