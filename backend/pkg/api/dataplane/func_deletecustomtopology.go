// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteCustomTopology Delete custom topology.
// @Summary Delete custom topology.
// @Description Delete custom topology.
// @Tags API.dataplane
// @Accept json
// @Produce json
// @Param Request body request.DeleteCustomTopologyRequest true "Delete Custom Topology Request"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/customtopology/delete [post]
func (h *handler) DeleteCustomTopology() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteCustomTopologyRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		err := h.dataplaneService.DeleteCustomTopology(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.DeleteCustomTopologyError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
