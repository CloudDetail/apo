// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

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
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.userService.GetUserTeam(req)
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
