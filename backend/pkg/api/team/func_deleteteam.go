// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// DeleteTeam Delete a team.
// @Summary Delete a team.
// @Description Delete a team.
// @Tags API.team
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param teamId formData int true "Team's id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/team/delete [post]
func (h *handler) DeleteTeam() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteTeamRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.teamService.DeleteTeam(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.DeleteTeamError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
