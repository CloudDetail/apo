// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetEndPointsData get the list of endpoints services
// @Summary get the list of endpoints services
// @Description get the list of endpoints services
// @Tags API.service
// @Accept application/json
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
// @Router /api/service/endpoints [post]
func (h *handler) GetEndPointsData() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetEndPointsDataRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		var data []response.ServiceEndPointsRes
		if allowed, err := h.dataService.CheckGroupPermission(c, req.GroupID); !allowed || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, []response.ServiceEndPointsRes{})
			return
		}

		data, err := h.serviceoverview.GetServicesEndPointData(c, req)
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
