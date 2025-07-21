// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// ListDataPlaneType
// @Summary
// @Description
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.listDataPlaneTypeResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/type/list [get]
func (h *handler) ListDataPlaneType() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.dataplaneService.ListDataPlaneType(c)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ListDataPlaneTypeError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
