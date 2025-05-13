// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// UpdateTeam Update team's information.
// @Summary Update team's information.
// @Description Update team's information.
// @Tags API.team
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.UpdateTeamRequest true "Request"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/team/update [post]
func (h *handler) UpdateTeam() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateTeamRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.teamService.UpdateTeam(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UpdateTeamError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
