// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetUserInfo Get user's info.
// @Summary Get user's info.
// @Description Get user's info.
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId query int false "User's id"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetUserInfoResponse
// @Failure 400 {object} code.Failure
// @Router /api/user/info [get]
func (h *handler) GetUserInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetUserInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		// TODO remove
		if req.UserID == 0 {
			req.UserID = c.UserID()
		}
		resp, err := h.userService.GetUserInfo(c, req.UserID)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetUserInfoError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
