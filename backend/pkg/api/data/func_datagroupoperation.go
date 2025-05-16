// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// DataGroupOperation Assign data groups to users or teams, or remove them from data groups.
// @Summary Assign data groups to users or teams, or remove them from data groups.
// @Description Assign data groups to users or teams, or remove them from data groups.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.DataGroupOperationRequest true "Request"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/data/group/operation [post]
func (h *handler) DataGroupOperation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DataGroupOperationRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataService.DataGroupOperation(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.AssignDataGroupError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
