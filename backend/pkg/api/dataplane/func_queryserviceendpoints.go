// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryServiceEndpoints Get service's endpoints.
// @Summary Get service's endpoints.
// @Description Get service's endpoints.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param cluster query string false "Query cluster name"
// @Success 200 {object} response.QueryServiceEndpointsResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/endpoints [get]
func (h *handler) QueryServiceEndpoints() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.QueryServiceEndpointsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		c.Payload(h.dataplaneService.GetServiceEndpoints(c, req))
	}
}
