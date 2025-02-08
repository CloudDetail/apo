// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

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
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.dataService.DataGroupOperation(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code)).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AssignDataGroupError,
					code.Text(code.AssignDataGroupError)).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
