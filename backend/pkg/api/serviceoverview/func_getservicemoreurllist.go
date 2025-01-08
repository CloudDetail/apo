// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/services/serviceoverview"
)

type getServiceMoreUrlListRequest struct {
	StartTime   int64  `form:"startTime" binding:"required"`                   // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield = StartTime"` // query end time
	Step        int64  `form:"step" binding:"required"`                        // step size
	ServiceName string `form:"serviceName" binding:"required"`                 // application name
	SortRule    int    `form:"sortRule" binding:"required"`                    // sort logic
}

// GetServiceMoreUrlList get more url list of services
// @Summary get more url list of services
// @Description get more url list of services
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param step query int64 true "step"
// @Param serviceName query string true "app name"
// @Param sortRule query int true "sort logic"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []response.ServiceDetail
// @Failure 400 {object} code.Failure
// @Router /api/service/moreUrl [get]
func (h *handler) GetServiceMoreUrlList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(getServiceMoreUrlListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		var startTime time.Time
		var endTime time.Time
		req.StartTime = req.StartTime / 1000000 // received microsecond-level startTime and endTime need to be converted to second-level first
		req.EndTime = req.EndTime / 1000000     // received microsecond-level startTime and endTime need to be converted to second-level first
		startTime = time.Unix(req.StartTime, 0)
		endTime = time.Unix(req.EndTime, 0)
		step := time.Duration(req.Step * 1000)
		//step := time.Minute
		serviceName := req.ServiceName
		sortRule := serviceoverview.SortType(req.SortRule)
		var res []response.ServiceDetail
		data, err := h.serviceoverview.GetServiceMoreUrl(startTime, endTime, step, serviceName, sortRule)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceMoreUrlListError,
				code.Text(code.GetServiceMoreUrlListError)).WithError(err),
			)
			return
		}
		if data != nil {
			res = data
		} else {
			res = []response.ServiceDetail{} // Make sure to return an empty array
		}

		c.Payload(res)
	}
}
