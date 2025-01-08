// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetSubjectDataGroup Get subject's assigned data group.
// @Summary Get subject's assigned data group.
// @Description Get subject's assigned data group.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param subjectId query int64 true "The id of authorized subject"
// @Param subjectType query string true "The type of authorized subject"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetSubjectDataGroupResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/sub/group [get]
func (h *handler) GetSubjectDataGroup() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSubjectDataGroupRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.dataService.GetSubjectDataGroup(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetDataGroupError,
				code.Text(code.GetDataGroupError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
