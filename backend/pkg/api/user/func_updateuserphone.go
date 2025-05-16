// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"net/http"
	"regexp"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// UpdateUserPhone Update phone number
// @Summary Update phone number
// @Description Update phone number
// @Tags API.user
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId formData int64 true "User's id"
// @Param phone formData string true "Phone number"
// @Param Authorization header string true "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/user/update/phone [post]
func (h *handler) UpdateUserPhone() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateUserPhoneRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if !phoneRegexp.MatchString(req.Phone) {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				nil,
			)
			return
		}

		err := h.userService.UpdateUserPhone(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UserUpdateError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}

var phoneRegexp = regexp.MustCompile("^1[3-9]\\d{9}$")
