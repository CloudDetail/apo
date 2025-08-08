// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryServices Get all services.
// @Summary Get all services.
// @Description Get all services.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param cluster query string false "Query cluster name"
// @Param currentPage query int false "Current page"
// @Param pageSize query int false "Page size"
// @Success 200 {object} response.QueryServicesResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/services [get]
func (h *handler) QueryServices() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.QueryServicesRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		if req.PageParam == nil {
			req.PageParam = &request.PageParam{
				CurrentPage: 1,
				PageSize:    10,
			}
		}
		c.Payload(h.dataplaneService.GetServices(c, req))
	}
}
