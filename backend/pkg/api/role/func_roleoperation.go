// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// RoleOperation Grant or revoke user's role.
// @Summary Grant or revoke user's role.
// @Description Grants permission to user
// @Tags API.role
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 ture "User's id"
// @Param roleList formData []int ture "The id list of role which user has." collectionFormat(multi)
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/role/operation [post]
func (h *handler) RoleOperation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.RoleOperationRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.roleService.RoleOperation(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					c.ErrMessage(vErr.Code),
				).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.UserGrantRoleError,
					c.ErrMessage(code.UserGrantRoleError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
