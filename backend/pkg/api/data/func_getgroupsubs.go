// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

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
// @Param dataGroupId query int64 true "data group's id"
// @Param subjectType query string false "subject type that you want to query"
// @Success 200 {object} response.GetGroupSubsResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/subs [get]
func (h *handler) GetGroupSubs() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetGroupSubsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.dataService.GetGroupSubs(req)
		if err != nil {
			c.HandleError(err, code.GetGroupSubsError)
			return
		}
		c.Payload(resp)
	}
}
