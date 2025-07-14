// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryAPPInfoTagValues Get app info tag values.
// @Summary Get app info tag values.
// @Description Get app info tag values.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.QueryAPPInfoTagValuesRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.QueryAPPInfoTagValuesResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/appinfo/tags/values [get]
func (h *handler) QueryAPPInfoTagValues() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.QueryAPPInfoTagValuesRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.dataplaneService.ListAPPInfoLabelValues(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.QueryAPPInfoValuesError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
