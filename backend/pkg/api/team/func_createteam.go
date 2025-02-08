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

// CreateTeam Creates a team.
// @Summary Creates a team.
// @Description Creates a team.
// @Tags API.team
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.CreateTeamRequest true "Request"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/team/create [post]
func (h *handler) CreateTeam() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateTeamRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.teamService.CreateTeam(req)
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
					code.CreateTeamError,
					code.Text(code.CreateTeamError)).WithError(err),
				)
			}
			return
		}
		c.Payload("ok")
	}
}
