// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// TeamOperation Assigns a user to teams or removes a user from teams.
// @Summary Assigns a user to teams or removes a user from teams.
// @Description Assigns a user to teams or removes a user from teams.
// @Tags API.team
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 true "User's id"
// @Param teamList formData []int64 false "The list of team id." collectionFormat(multi)
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/team/operation [post]
func (h *handler) TeamOperation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.TeamOperationRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.teamService.TeamOperation(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.AssignToTeamError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
