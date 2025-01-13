// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// CountK8sEvents get K8s events
// @Summary get K8s events
// @Description get K8s events
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetK8sEventsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/k8s/events/count [get]
func (h *handler) CountK8sEvents() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetK8sEventsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.CountK8sEvents(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetK8sEventError,
				code.Text(code.GetK8sEventError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
