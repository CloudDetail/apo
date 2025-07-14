// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryAPPInfoTags Get app info tags.
// @Summary Get app info tags.
// @Description Get app info tags.
// @Tags API.dataplane
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.QueryAPPInfoTagsRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.QueryAPPInfoTagsResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/appinfo/tags [get]
func (h *handler) QueryAPPInfoTags() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.QueryAPPInfoTagsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.dataplaneService.ListAPPInfoLabelsKeys(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.QueryAPPInfoTagsError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
