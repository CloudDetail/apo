// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// UpdateDataGroup Updates data group.
// @Summary Updates data group.
// @Description Updates data group.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.UpdateDataGroupRequest true "Request"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/data/group/update [post]
func (h *handler) UpdateDataGroup() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateDataGroupRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataService.UpdateDataGroup(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UpdateDataGroupError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
