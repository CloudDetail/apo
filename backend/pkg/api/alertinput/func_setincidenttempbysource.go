// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// SetIncidentTempBySource SetIncidentTempBySource
// @Summary
// @Description
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.SetIncidentTempBySourceRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/incident/temp/set [post]
func (h *handler) SetIncidentTempBySource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetIncidentTempBySourceRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.inputService.SetIncidentTempBySource(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				code.SetIncidentTempBySourceError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
