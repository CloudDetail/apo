// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// CleanDataSourceServices clean relate service by dataSource.
// @Summary clean relate service by dataSource.
// @Description clean relate service by dataSource.
// @Tags API.dataplane
// @Accept json
// @Produce json
// @Param Request body request.CleanDataSourceServicesRequest true "Request information"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/servicename/cleanDataSourceServices [post]
func (h *handler) CleanDataSourceServices() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CleanDataSourceServicesRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if err := h.dataplaneService.CleanRelateServiceByDataSource(c, req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CleanRelateServiceError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
