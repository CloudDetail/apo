// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServiceEndPointList get the list of service EndPoint
// @Summary get the list of service EndPoint
// @Description get the list of service EndPoint
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string false "Query service name"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []string
// @Failure 400 {object} code.Failure
// @Router /api/service/endpoint/list [post]
func (h *handler) GetServiceEndPointList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceEndPointListRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allowed, err := h.dataService.CheckServicesPermission(c, req.ServiceName); !allowed || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, []string{})
			return
		}

		resp, err := h.serviceInfoService.GetServiceEndPointList(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetServiceEndPointListError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
