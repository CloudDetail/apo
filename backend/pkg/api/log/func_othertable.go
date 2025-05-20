// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// OtherTable get external log table information
// @Summary get external log table
// @Description get external log table
// @Tags API.log
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.OtherTableRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.OtherTableResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/other [get]
func (h *handler) OtherTable() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.OtherTableRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp, err := h.logService.OtherTable(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetAllOtherLogTableError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
