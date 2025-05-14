// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetUserTeam Get user's team.
// @Summary Get user's team.
// @Description Get user's team.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId query int64 true "User's is"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetUserTeamResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/team [get]
func (h *handler) GetUserTeam() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetUserTeamRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.userService.GetUserTeam(c, req)
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
