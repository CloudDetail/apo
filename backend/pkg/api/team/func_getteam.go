// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetTeam Get teams.
// @Summary Get teams.
// @Description Get teams.
// @Tags API.team
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param teamName query string false "Team's name"
// @Param featureList query []int false "The list of feature's id" collectionFormat(multi)
// @Param dataGroupList query []int false "The list of data group's id" collectionFormat(multi)
// @Param currentPage query int false "Current page"
// @Param pageSize query int false "Page size"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetTeamResponse
// @Failure 400 {object} code.Failure
// @Router /api/team [get]
func (h *handler) GetTeam() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTeamRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if req.PageParam == nil {
			req.PageParam = &request.PageParam{
				CurrentPage: 1,
				PageSize:    10,
			}
		}

		resp, err := h.teamService.GetTeamList(req)
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
