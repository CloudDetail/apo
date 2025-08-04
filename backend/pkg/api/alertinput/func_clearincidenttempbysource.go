// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// ClearIncidentTempBySource ClearIncidentTempBySource
// @Summary
// @Description
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.ClearIncidentTempBySourceRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/incident/temp/clear [get]
func (h *handler) ClearIncidentTempBySource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ClearIncidentTempBySourceRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.inputService.ClearIncidentTempBySource(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				code.ClearIncidentTempBySourceError,
				err,
			)
			return
		}

		c.Payload("ok")
	}
}
