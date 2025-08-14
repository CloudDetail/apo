// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

// CheckDataSource check datasource is valid.
// @Summary check datasource is valid.
// @Description check datasource is valid.
// @Tags API.dataplane
// @Accept json
// @Produce json
// @Param Request body model.CheckDataSourceRequest true "Request"
// @Success 200 {object} model.CheckDataSourceResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/checkDataSource [post]
func (h *handler) CheckDataSource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(model.CheckDataSourceRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		c.Payload(h.dataplaneService.CheckDataSource(c, req))
	}
}
