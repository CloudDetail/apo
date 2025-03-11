// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/internal/model/response"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// Delete Delete xx
// @Summary delete xx
// @Description delete xx
// @Tags API.mock
// @Accept json
// @Produce json
// @Param id path uint true "Id"
// @Success 200 {object} response.DeleteResponse
// @Failure 400 {object} code.Failure
// @Router /api/mock/{id} [delete]
func (h *handler) Delete() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteRequest)
		if err := c.ShouldBindURI(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.mockService.Delete(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.MockDetailError,
				c.ErrMessage(code.MockDetailError)).WithError(err),
			)
			return
		}
		resp := new(response.DeleteResponse)
		resp.Id = req.Id
		c.Payload(resp)
	}
}
