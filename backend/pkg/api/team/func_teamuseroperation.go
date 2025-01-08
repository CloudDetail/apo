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

// TeamUserOperation Assigns users to a team or remove users from a team.
// @Summary Assigns users to a team or remove users from a team.
// @Description Assigns users to a team or remove users from a team.
// @Tags API.team
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param teamId formData int64 true "Team's id"
// @Param userList formData []int64 false "The list of users' id." collectionFormat(multi)
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/team/user/operation [post]
func (h *handler) TeamUserOperation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AssignToTeamRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.teamService.TeamUserOperation(req)
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
					code.AssignToTeamError,
					code.Text(code.AssignToTeamError)).WithError(err),
				)
			}
			return
		}
		c.Payload("ok")
	}
}
