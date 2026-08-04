// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alertanalyze

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetAlertEvents get alarm events
// @Summary get alarm events
// @Description get alarm events
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.GetAlertEventsSvcReq true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetAlertEventsSvcResp
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alert/events [post]
func (h *handler) GetAlertEvents() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertEventsSvcReq)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.alertanalyzeService.GetAlertEventsSvc(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAlertEventsError,
				err,
			)
			return
		}
		if resp == nil {
			resp = &response.GetAlertEventsSvcResp{
				Pagination: &model.Pagination{
					Total: 0,
				},
				Events: []response.ImpactAlertEvent{},
			}
		}
		c.Payload(resp)
	}
}
