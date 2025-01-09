// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// Create Create/Edit xx
// @Summary create/edit xx
// @Description create/edit xx
// @Tags API.mock
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param name formData string true "name"
// @Success 200 {object} response.CreateResponse
// @Failure 400 {object} code.Failure
// @Router /api/mock [post]
func (h *handler) Create() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.mockService.Create(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.MockCreateError,
				code.Text(code.MockCreateError)).WithError(err),
			)
			return
		}

		c.Payload(resp)
	}
}
