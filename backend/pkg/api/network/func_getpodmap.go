// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package deepflow

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetPodMap query pod network call topology and call metrics
// @Summary query pod network call topology and call metrics
// @Description query pod network call topology and call metrics
// @Tags API.Network
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "Start time, in microseconds"
// @Param endTime query int64 true "End time, in microseconds"
// @Param namespace query string false "Namespace to query, if the value is empty, query all"
// @Param workload query string false "Workload to be queried. If the value is empty, all of them will be queried"
// @Success 200 {object} response.PodMapResponse
// @Failure 400 {object} code.Failure
// @Router /api/network/podmap [get]
func (h *handler) GetPodMap() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.PodMapRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		userID := c.UserID()
		err := h.dataService.CheckDatasourcePermission(userID, 0, &req.Namespace, nil, "")
		if err != nil {
			c.AbortWithPermissionError(err, code.AuthError, new(response.PodMapResponse))
			return
		}
		resp, err := h.networkService.GetPodMap(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ServerError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
