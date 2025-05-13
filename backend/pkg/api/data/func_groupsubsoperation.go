// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GroupSubsOperation Manage group's assigned subject.
// @Summary Manage group's assigned subject.
// @Description Manage group's assigned subject.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GroupSubsOperationRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/data/subs/operation [post]
func (h *handler) GroupSubsOperation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GroupSubsOperationRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataService.GroupSubsOperation(req)
		if err != nil {
			c.AbortWithPermissionError(err, code.AssignDataGroupError, nil)
			return
		}

		c.Payload("ok")
	}
}
