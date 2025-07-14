// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServiceInstanceOptions get the drop-down list of service instances
// @Summary get the drop-down list of service instances
// @Description get the drop-down list of service instances
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} map[string]model.ServiceInstance
// @Failure 400 {object} code.Failure
// @Router /api/service/instance/options [post]
func (h *handler) GetServiceInstanceOptions() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceInstanceOptionsRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allowed, err := h.dataService.CheckServicesPermission(c, req.ServiceName); !allowed || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, make(map[string]*model.ServiceInstance))
			return
		}

		resp, err := h.serviceInfoService.GetServiceInstanceOptions(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetServiceInstanceOptionsError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
