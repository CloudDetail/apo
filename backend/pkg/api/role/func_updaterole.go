// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// UpdateRole Update role's name and permission.
// @Summary Update role's name and permission.
// @Description Update role's name and permission.
// @Tags API.role
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param roleId formData int true "Role's id."
// @Param roleName formData string true "Role's name"
// @Param description formData string false "The description of role."
// @Param permissionList formData []int false "Role's feature permission id list." collectionFormat(multi)
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/role/update [post]
func (h *handler) UpdateRole() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateRoleRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.roleService.UpdateRole(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UpdateRoleError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
