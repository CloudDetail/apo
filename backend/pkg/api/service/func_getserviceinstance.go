// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"
	"time"

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
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param step query int64 true "step"
// @Param serviceName query string true "app name"
// @Param endpoint query string false "endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.InstancesRes
// @Failure 400 {object} code.Failure
// @Router /api/service/instances [get]
func (h *handler) GetServiceInstance() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceInstanceRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		userID := c.UserID()
		err := h.dataService.CheckDatasourcePermission(userID, 0, nil, &req.ServiceName, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.AbortWithPermissionError(err, code.AuthError, &response.InstancesRes{
				Status: model.STATUS_NORMAL,
				Data:   []response.InstanceData{},
			})
			return
		}
		var startTime time.Time
		var endTime time.Time
		req.StartTime = req.StartTime / 1000000 // received microsecond-level startTime and endTime need to be converted to second-level first
		req.EndTime = req.EndTime / 1000000     // received microsecond-level startTime and endTime need to be converted to second-level first
		startTime = time.Unix(req.StartTime, 0)
		endTime = time.Unix(req.EndTime, 0)
		step := time.Duration(req.Step * 1000)
		serviceName := req.ServiceName
		endpoint := req.Endpoint
		data, err := h.serviceInfoService.GetInstancesNew(startTime, endTime, step, serviceName, endpoint)
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
