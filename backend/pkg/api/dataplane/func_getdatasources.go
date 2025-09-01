// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetDatasources
// @Summary
// @Description
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param cluster query string false "Cluseter id"
// @Param startTime query int64 true "Start time"
// @Param endTime query int64 true "End time"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetDatasourcesResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/datasources [get]
func (h *handler) GetDatasources() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDatasourcesRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.dataplaneService.GetDatasources(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetDatasourcesError,
				err,
			)
		}
		c.Payload(resp)
	}
}
