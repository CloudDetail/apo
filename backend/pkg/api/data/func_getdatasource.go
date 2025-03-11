// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetDatasource Gets all datasource.
// @Summary Gets all datasource.
// @Description Gets all datasource.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetDatasourceResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/datasource [get]
func (h *handler) GetDatasource() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.dataService.GetDataSource()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetDatasourceError,
				c.ErrMessage(code.GetDatasourceError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
