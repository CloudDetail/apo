// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// ListCustomTopology List custom topology.
// @Summary List custom topology.
// @Description List custom topology.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Success 200 {object} response.ListCustomTopologyResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/customtopology/list [get]
func (h *handler) ListCustomTopology() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ListCustomTopologyRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.dataplaneService.ListCustomTopology(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ListCustomTopologyError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
