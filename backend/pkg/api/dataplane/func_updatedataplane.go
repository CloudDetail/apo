// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UpdateDataPlane
// @Summary
// @Description
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.UpdateDataPlaneRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/update [post]
func (h *handler) UpdateDataPlane() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateDataPlaneRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataplaneService.UpdateDataPlane(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UpdateDataPlaneError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
