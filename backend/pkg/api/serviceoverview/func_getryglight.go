// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

// GetRYGLight get traffic light results
// @Summary get traffic light results
// @Description get traffic light results
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param serviceName query string false "Service name"
// @Param endpointName query string false "interface name"
// @Param namespace query string false "namespace"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.ServiceRYGLightRes
// @Failure 400 {object} code.Failure
// @Router /api/service/ryglight [get]
func (h *handler) GetRYGLight() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetRygLightRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, 0, &req.Namespace, &req.ServiceName, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.HandleError(err, code.AuthError)
			return
		}
		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)
		filter := serviceoverview.EndpointsFilter{
			ContainsSvcName:      req.ServiceName,
			ContainsEndpointName: req.EndpointName,
			Namespace:            req.Namespace,
		}

		resp, err := h.serviceoverview.GetServicesRYGLightStatus(startTime, endTime, filter)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceMoreUrlListError,
				code.Text(code.GetServiceMoreUrlListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
