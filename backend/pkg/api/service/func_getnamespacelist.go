// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetNamespaceList Get monitored namespaces.
// @Summary Get monitored namespaces.
// @Description Get monitored namespaces.
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "start time"
// @Param endTime query int64 true "end time"
// @Success 200 {object} response.GetServiceNamespaceListResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/namespace/list [get]
func (h *handler) GetNamespaceList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceNamespaceListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.serviceInfoService.GetServiceNamespaceList(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetNamespaceListError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
