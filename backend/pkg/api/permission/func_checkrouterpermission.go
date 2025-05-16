// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package permission

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// CheckRouterPermission Check a router is authorized to view.
// @Summary Check a router is authorized to view.
// @Description Check a router is authorized to view.
// @Tags API.permission
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param router query string true "Router"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.CheckRouterPermissionResponse
// @Failure 400 {object} code.Failure
// @Router /api/permission/router [get]
func (h *handler) CheckRouterPermission() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CheckRouterPermissionRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		userID := c.UserID()
		resp, err := h.permissionService.CheckRouterPermission(userID, req.Router)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CheckRouterError,
				err,
			)
		}
		c.Payload(resp)
	}
}
