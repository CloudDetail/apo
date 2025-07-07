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
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

// GetAlertEventsSample get sampling alarm events
// @Summary get sampling alarm events
// @Description get sampling alarm events
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param source query string false "Query the alarm source"
// @Param group query string false "Query alarm type"
// @Param name query string false "Query alarm name"
// @Param id query string false "Query alarm ID"
// @Param status query string false "Query alarm status"
// @Param sampleCount query int false "Number of sample alarms"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetAlertEventsSampleResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/alert/sample/events [post]
func (h *handler) GetAlertEventsSample() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertEventsSampleRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		if len(req.Service) > 0 {
			req.Services = append(req.Services, req.Service)
		}

		if allowed, err := h.dataService.CheckGroupPermission(c, req.GroupID); !allowed || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, &response.GetAlertEventsSampleResponse{
				EventMap: map[string]map[string][]clickhouse.AlertEventSample{},
				Status:   model.STATUS_NORMAL,
			})
			return
		}

		resp, err := h.serviceInfoService.GetAlertEventsSample(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertEventsError,
				err,
			)
			return
		}
		if resp == nil {
			resp = &response.GetAlertEventsSampleResponse{
				EventMap: map[string]map[string][]clickhouse.AlertEventSample{},
			}
		}
		c.Payload(resp)
	}
}
