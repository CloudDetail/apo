// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// CreateDataGroup Create a data group.
// @Summary Create a data group.
// @Description Create a data group.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.CreateDataGroupRequest true "Request"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/data/group/create [post]
func (h *handler) CreateDataGroup() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateDataGroupRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataService.CreateDataGroup(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CreateDataGroupError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
