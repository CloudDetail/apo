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

// CountK8sEvents get K8s events
// @Summary get K8s events
// @Description get K8s events
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetK8sEventsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/k8s/events/count [post]
func (h *handler) CountK8sEvents() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetK8sEventsRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allow, err := h.dataService.CheckGroupPermission(c, req.GroupID); !allow || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, &response.GetK8sEventsResponse{
				Status:  model.STATUS_NORMAL,
				Reasons: []string{},
				Data:    make(map[string]*response.K8sEventStatistics),
			})
		}

		resp, err := h.serviceInfoService.CountK8sEvents(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetK8sEventError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
