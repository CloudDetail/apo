// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetDataGroup Get data group.
// @Summary Get data group.
// @Description Get data group.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetDataGroupRequest false "Request"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetDataGroupResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/group [post]
func (h *handler) GetDataGroup() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDataGroupRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		if req.PageParam == nil {
			req.PageParam = &request.PageParam{
				CurrentPage: 1,
				PageSize:    10,
			}
		}

		resp, err := h.dataService.GetDataGroup(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetDataGroupError,
				c.ErrMessage(code.GetDataGroupError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
