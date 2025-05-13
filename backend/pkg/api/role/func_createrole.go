// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// CreateRole Creates a role.
// @Summary Creates a role.
// @Description Creates a role.
// @Tags API.role
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param roleName formData string true "Role's name"
// @Param description formData string false "The description of role."
// @Param permissionList formData []int false "Role's initial feature permission id list." collectionFormat(multi)
// @Param userList formData []int false "The id of users which will be granted the role." collectionFormat(multi)
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/role/create [post]
func (h *handler) CreateRole() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateRoleRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.roleService.CreateRole(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CreateRoleError,
				err,
			)
			return
		}

		c.Payload("ok")
	}
}
