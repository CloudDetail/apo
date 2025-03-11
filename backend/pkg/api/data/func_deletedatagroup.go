// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// DeleteDataGroup Delete the data group.
// @Summary Delete the data group.
// @Description Delete the data group.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param groupId formData int64 true "Data group's id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/data/group/delete [post]
func (h *handler) DeleteDataGroup() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteDataGroupRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.dataService.DeleteDataGroup(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					c.ErrMessage(vErr.Code)).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.DeleteDataGroupError,
					c.ErrMessage(code.DeleteDataGroupError)).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
