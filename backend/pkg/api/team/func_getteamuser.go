// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetTeamUser Get team's users.
// @Summary Get team's users.
// @Description Get team's users.
// @Tags API.team
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param teamId query int64 true "Team's id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetTeamUserResponse
// @Failure 400 {object} code.Failure
// @Router /api/team/user [get]
func (h *handler) GetTeamUser() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTeamUserRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.teamService.GetTeamUser(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetTeamError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
