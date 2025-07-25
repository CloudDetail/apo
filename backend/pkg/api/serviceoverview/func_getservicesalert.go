// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServicesAlert get the log alarm and status light information of the Service
// @Summary get the log alarm and status light information of the Service
// @Description get the log alarm and status light information of the Service
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param step query int64 true "step"
// @Param serviceNames query []string true "application name" collectionFormat(multi)
// @Param returnData query []string false "return data content" collectionFormat(multi)
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.ServiceAlertRes
// @Failure 400 {object} code.Failure
// @Router /api/service/servicesAlert [post]
func (h *handler) GetServicesAlert() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceAlertRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		// userID := c.UserID()
		// err := h.dataService.CheckDatasourcePermission(c, userID, 0, nil, &req.ServiceNames, model.DATASOURCE_CATEGORY_APM)
		// if err != nil {
		// 	c.AbortWithPermissionError(err, code.AuthError, []response.ServiceAlertRes{})
		// 	return
		// }

		var startTime time.Time
		var endTime time.Time
		req.StartTime = req.StartTime / 1000000 // received microsecond-level startTime and endTime need to be converted to second-level first
		req.EndTime = req.EndTime / 1000000     // received microsecond-level startTime and endTime need to be converted to second-level first
		startTime = time.Unix(req.StartTime, 0)
		endTime = time.Unix(req.EndTime, 0)
		step := time.Duration(req.Step * 1000)
		returnData := req.ReturnData

		if len(req.ServiceName) > 0 {
			req.ServiceNames = append(req.ServiceNames, req.ServiceName)
		}

		var resp []response.ServiceAlertRes
		data, err := h.serviceoverview.GetServicesAlert(c, 0, req.ClusterIDs, startTime, endTime, step, req.ServiceNames, returnData)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetServicesAlertError,
				err,
			)
			return
		}
		if data != nil {
			resp = data
		} else {
			resp = []response.ServiceAlertRes{}
		}

		c.Payload(resp)
	}
}
