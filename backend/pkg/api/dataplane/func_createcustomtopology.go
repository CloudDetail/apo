// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// CreateCustomTopology Create custom topology.
// @Summary Create custom topology.
// @Description Create custom topology.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param clusterId query string true "cluster Id"
// @Param leftNode query string true "parent node name"
// @Param leftType query string true "parent node type"
// @Param rightNode query string true "child node name"
// @Param rightType query string true "child node type"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/customtopology/create [post]
func (h *handler) CreateCustomTopology() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateCustomTopologyRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		err := h.dataplaneService.CreateCustomTopology(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CreateCustomTopologyError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
