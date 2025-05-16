// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetServiceREDCharts Get services' red charts.
// @Summary Get services' red charts.
// @Description Get services' red charts.
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param request body request.GetServiceREDChartsRequest true "request"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetServiceREDChartsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/redcharts [post]
func (h *handler) GetServiceREDCharts() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceREDChartsRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		userID := c.UserID()
		err := h.dataService.CheckDatasourcePermission(c, userID, 0, nil, &req.ServiceList, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.AbortWithPermissionError(err, code.AuthError, response.GetServiceREDChartsResponse{})
		}
		resp, err := h.serviceInfoService.GetServiceREDCharts(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetServiceListError,
				err,
			)
			return
		}

		c.Payload(resp)
	}
}
