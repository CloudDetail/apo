// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// ListDataPlane
// @Summary
// @Description
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// TODO The following request parameter types and response types must be changed according to actual requirements.
// @Param Request body request.listDataPlaneRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.listDataPlaneResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/list [get]
func (h *handler) ListDataPlane() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.dataplaneService.ListDataPlane(c)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ServerError, // TODO err code
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
