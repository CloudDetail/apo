// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// PermissionOperation Grant or revoke user's permission(feature).
// @Summary Grant or revoke user's permission(feature).
// @Description Grant or revoke user's permission(feature).
// @Tags API.permission
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param subjectId formData int64 true "authorization principal id"
// @Param subjectType formData string true "Authorization principal type: 'role','user','team '"
// @Param type formData string true "Authorization type: 'feature','data '"
// @Param permissionList formData []int false "list of permission ids" collectionFormat(multi)
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/permission/operation [post]
func (h *handler) PermissionOperation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.PermissionOperationRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.userService.PermissionOperation(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code),
				).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.UserGrantPermissionError,
					code.Text(code.UserGrantPermissionError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
