// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// CreateDataPlane
// @Summary
// @Description
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.CreateDataPlaneRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/create [post]
func (h *handler) CreateDataPlane() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateDataPlaneRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataplaneService.CreateDataPlane(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CreateDataPlaneError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
