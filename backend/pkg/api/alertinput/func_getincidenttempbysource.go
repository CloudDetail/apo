// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetIncidentTempBySource GetIncidentTempBySource
// @Summary
// @Description
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetIncidentTempBySourceRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetIncidentTempBySourceResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/incident/temp/get [get]
func (h *handler) GetIncidentTempBySource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetIncidentTempBySourceRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.inputService.GetIncidentTempBySource(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				code.GetIncidentTempBySourceError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
