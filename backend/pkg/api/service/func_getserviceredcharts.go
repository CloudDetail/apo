// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServiceREDCharts Get services' red charts.
// @Summary Get services' red charts.
// @Description Get services' red charts.
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "start time"
// @Param endTime query int64 true "end time"
// @Param step query int64 true "step"
// @Param serviceList query []string true "service list" collectionFormat(multi)
// @Param endpointList query []string true "endpoint list" collectionFormat(multi)
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetServiceREDChartsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/redcharts [get]
func (h *handler) GetServiceREDCharts() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceREDChartsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.GetServiceREDChartsError),
			))
			return
		}

		resp, err := h.serviceInfoService.GetServiceREDCharts(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceListError,
				c.ErrMessage(code.GetServiceREDChartsError)).WithError(err))
			return
		}
		
		c.Payload(resp)
	}
}
