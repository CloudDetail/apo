// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetUserDatasource Get datasource that user authorized to view.
// @Summary Get datasource that user authorized to view.
// @Description Get datasource that user authorized to view.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param category query string false "apm or normal"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.getUserDatasourceResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/user [get]
func (h *handler) GetUserDatasource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetUserDatasourceRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		c.Payload("ok")
	}
}
