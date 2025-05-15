// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

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
// @Param category query string false "apm or normal, return all if is empty"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetSubjectDataGroupResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/sub/group [get]
func (h *handler) GetSubjectDataGroup() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSubjectDataGroupRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.dataService.GetSubjectDataGroup(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetDataGroupError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
