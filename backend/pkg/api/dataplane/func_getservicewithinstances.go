// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GGetServiceWithInstances
// @Summary
// @Description
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param cluster query string false "Cluseter id"
// @Param startTime query int64 true "Start time"
// @Param endTime query int64 true "End time"
// @Param services query []string false "Services" collectionFormat(multi)
// @Param currentPage query int false "Current page"
// @Param pageSize query int false "Page size"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetServiceWithInstancesResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/servicesWithInstances [get]
func (h *handler) GetServiceWithInstances() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceWithInstancesRequest)
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
		resp, err := h.dataplaneService.GetServiceWithInstances(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetServiceWithInstancesError,
				err,
			)
		}
		c.Payload(resp)
	}
}
