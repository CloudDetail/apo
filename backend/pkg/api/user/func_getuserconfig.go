// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetUserConfig Gets user's menu config and which route can access.
// @Summary Gets user's menu config and which route can access.
// @Description Get user's menu config.
// @Tags API.permission
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId query int64 true "用户id"
// @Param language query string false "language"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetUserConfigResponse
// @Failure 400 {object} code.Failure
// @Router /api/permission/config [get]
func (h *handler) GetUserConfig() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetUserConfigRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		if len(req.Language) == 0 {
			req.Language = model.TRANSLATION_ZH
		}

		resp, err := h.userService.GetUserConfig(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetMenuConfigError,
				code.Text(code.GetMenuConfigError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
