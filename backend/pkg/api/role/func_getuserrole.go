// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package role

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetUserRole Get user's role.
// @Summary Get user's role.
// @Description Get user's role.
// @Tags API.role
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId query int64 true "User's id"
// @Success 200 {object} response.GetUserRoleResponse
// @Failure 400 {object} code.Failure
// @Router /api/role/user [get]
func (h *handler) GetUserRole() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetUserRoleRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.roleService.GetUserRole(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UserGetRolesERROR,
				code.Text(code.UserGetRolesERROR)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
