// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

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
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.teamService.GetTeamUser(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code)).WithError(err),
				)
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.GetTeamError,
					code.Text(code.GetTeamError)).WithError(err),
				)
			}
			return
		}
		c.Payload(resp)
	}
}
