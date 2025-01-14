// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"
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
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err))
			return
		}
		// TODO remove
		// TODO remove
		// TODO remove
		if req.UserID == 0 {
			req.UserID = middleware.GetContextUserID(c)
		}
		resp, err := h.userService.GetUserInfo(req.UserID)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code)).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.GetUserInfoError,
					code.Text(code.GetUserInfoError)).WithError(err))
			}
			return
		}
		c.Payload(resp)
	}
}
