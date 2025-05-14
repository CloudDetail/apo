// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetGroupSubs Get group's assigned subjects.
// @Summary Get group's assigned subjects.
// @Description Get group's assigned subjects.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Param groupId query int64 true "data group's id"
// @Param subjectType query string false "subject type that you want to query"
// @Success 200 {object} response.GetGroupSubsResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/subs [get]
func (h *handler) GetGroupSubs() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetGroupSubsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.dataService.GetGroupSubs(c, req)
		if err != nil {
			c.AbortWithPermissionError(err, code.GetGroupSubsError, nil)
			return
		}
		c.Payload(resp)
	}
}
