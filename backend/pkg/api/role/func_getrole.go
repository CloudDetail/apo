// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetRole Gets all roles.
// @Summary Gets all roles.
// @Description Gets all roles.
// @Tags API.role
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} response.GetRoleResponse
// @Failure 400 {object} code.Failure
// @Router /api/role/roles [get]
func (h *handler) GetRole() core.HandlerFunc {
	return func(c core.Context) {

		resp, err := h.roleService.GetRoles()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserGetRolesERROR,
				c.ErrMessage(code.UserGetRolesERROR),
			))
		}
		c.Payload(resp)
	}
}
