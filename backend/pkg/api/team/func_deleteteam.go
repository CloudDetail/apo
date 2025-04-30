// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"
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
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.teamService.DeleteTeam(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					c.ErrMessage(vErr.Code)).WithError(err),
				)
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.DeleteTeamError,
					c.ErrMessage(code.DeleteTeamError)).WithError(err),
				)
			}
			return
		}
		c.Payload("ok")
	}
}
