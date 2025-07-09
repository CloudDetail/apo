// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetGroupDatasource Get group's datasource.
// @Summary Get group's datasource.
// @Description Get group's datasource.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param groupId query int64 false "Data group's id"
// @Param category query string false "apm or normal, return all when empty"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetGroupDatasourceResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/group/data [get]
func (h *handler) GetGroupDatasource() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetGroupDatasourceRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allowed, err := h.dataService.CheckGroupPermission(c, req.GroupID); !allowed || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, response.GetGroupDatasourceResponse{})
			return
		}

		resp, err := h.dataService.GetGroupDatasource(c, req)
		if err != nil {
			c.AbortWithPermissionError(err, code.GetGroupDatasourceError, nil)
			return
		}
		c.Payload(resp)
	}
}
