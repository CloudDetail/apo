// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetEndPointsData get the list of endpoints services
// @Summary get the list of endpoints services
// @Description get the list of endpoints services
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param groupId query int64 false "data group id"
// @Param startTime query int64 true "start time"
// @Param endTime query int64 true "end time"
// @Param step query int64 true "step"
// @Param serviceName query []string false "service name" collectionFormat(multi)
// @Param namespace query []string false "namespace" collectionFormat(multi)
// @Param endpointName query []string false "endpoint" collectionFormat(multi)
// @Param sortRule query int true "sort rule"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.ServiceEndPointsRes
// @Failure 400 {object} code.Failure
// @Router /api/service/endpoints [get]
func (h *handler) GetEndPointsData() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetEndPointsDataRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		var data []response.ServiceEndPointsRes
		var startTime time.Time
		var endTime time.Time
		startTime = time.UnixMicro(req.StartTime)
		endTime = time.UnixMicro(req.EndTime)
		step := time.Duration(req.Step * 1000)

		userID := c.UserID()
		err := h.dataService.CheckDatasourcePermission(c, userID, req.GroupID, &req.Namespace, &req.ServiceName, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.AbortWithPermissionError(err, code.AuthError, []response.ServiceEndPointsRes{})
			return
		}

		filter := serviceoverview.EndpointsFilter{
			MultiService:   req.ServiceName,
			MultiEndpoint:  req.EndpointName,
			MultiNamespace: req.Namespace,
		}
		data, err = h.serviceoverview.GetServicesEndPointData(c, startTime, endTime, step, filter, req.SortRule)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetTop3UrlListError,
				err,
			)
			return
		}

		c.Payload(data)
	}
}
