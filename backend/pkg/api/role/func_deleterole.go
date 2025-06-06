// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

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
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.roleService.DeleteRole(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.DeleteRoleError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
