// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetServiceInstance get the URL instance corresponding to the service.
// @Summary get the URL instance corresponding to the service
// @Description get the URL instance corresponding to the service
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param step query int64 true "step"
// @Param serviceName query string true "app name"
// @Param endpoint query string false "endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.InstancesRes
// @Failure 400 {object} code.Failure
// @Router /api/service/instances [post]
func (h *handler) GetServiceInstance() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceInstanceRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allow, err := h.dataService.CheckGroupPermission(c, req.GroupID); !allow || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, &response.InstancesRes{
				Status: model.STATUS_NORMAL,
				Data:   []response.InstanceData{},
			})
			return
		}

		data, err := h.serviceInfoService.GetInstancesNew(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetOverviewServiceInstanceListError,
				err,
			)
			return
		}
		c.Payload(data)
	}
}
