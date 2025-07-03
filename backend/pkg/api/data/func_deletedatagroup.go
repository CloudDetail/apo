// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// DeleteDataGroupV2 Delete the data group.
// @Summary Delete the data group.
// @Description Delete the data group.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param groupId formData int64 true "Data group's id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/v2/data/group/delete [post]
func (h *handler) DeleteDataGroupV2() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteDataGroupRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataService.DeleteDataGroupV2(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.DeleteDataGroupError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
