// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// CheckDataPlane
// @Summary
// @Description
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} response.CheckMinimalDataPlaneResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/data/check [get]
func (h *handler) CheckDataPlane() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.dataplaneService.CheckMinimalDataPlaneExist(c)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CheckMinimalDataPlaneExistError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
