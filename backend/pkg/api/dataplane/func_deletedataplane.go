// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteDataPlane
// @Summary
// @Description
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.deleteDataPlaneRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/delete [post]
func (h *handler) DeleteDataPlane() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteDataPlaneRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		// TODO replace with Service call
		err := h.dataplaneService.DeleteDataPlane(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ServerError, // TODO DeleteDataPlaneError
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
