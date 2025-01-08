// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteOtherTable remove external log table
// @Summary remove external log table
// @Description remove external log table
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.DeleteOtherTableRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.DeleteOtherTableResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/other/delete [delete]
func (h *handler) DeleteOtherTable() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteOtherTableRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.DeleteOtherTable(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.DeleteOtherLogTableError,
				code.Text(code.DeleteOtherLogTableError)+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
