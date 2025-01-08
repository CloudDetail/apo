// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// DeleteRole Delete a role.
// @Summary Delete a role.
// @Description Delete a role.
// @Tags API.role
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param roleId formData int true "Role's id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @router /api/role/delete [post]
func (h *handler) DeleteRole() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteRoleRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.roleService.DeleteRole(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code)).WithError(err),
				)
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.DeleteRoleError,
					code.Text(code.DeleteRoleError)).WithError(err),
				)
			}
			return
		}
		c.Payload("ok")
	}
}
