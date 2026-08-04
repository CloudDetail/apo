// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetAlertEvents get alarm events
// @Summary get alarm events
// @Description get alarm events
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param Request body request.GetAlertEventsRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetAlertEventsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/alert/events [post]
func (h *handler) GetAlertEvents() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertEventsRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		// TODO Compatible with old APIs
		{
			if len(req.Service) > 0 {
				req.Services = append(req.Services, req.Service)
			}
			if len(req.Endpoint) > 0 {
				req.Endpoints = append(req.Endpoints, req.Endpoint)
			}
		}

		resp, err := h.serviceInfoService.GetAlertEvents(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertEventsError,
				err,
			)
			return
		}
		if resp == nil {
			resp = &response.GetAlertEventsResponse{
				TotalCount: 0,
				EventList:  []alert.AlertEvent{},
			}
		}
		c.Payload(resp)
	}
}
