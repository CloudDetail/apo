// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// Detail xx Details
// @Summary xx details
// @Description xx details
// @Tags API.mock
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} response.DetailResponse
// @Failure 400 {object} code.Failure
// @Router /api/mock/{id} [get]
func (h *handler) Detail() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DetailRequest)
		if err := c.ShouldBindURI(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.mockService.Detail(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.MockDetailError,
				c.ErrMessage(code.MockDetailError)).WithError(err),
			)
			return
		}

		c.Payload(resp)
	}
}
