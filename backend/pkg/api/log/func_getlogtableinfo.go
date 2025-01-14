// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetLogTableInfo
// @Summary get log table information
// @Description get log table information
// @Tags API.log
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.LogTableInfoRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.LogTableInfoResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/table [get]
func (h *handler) GetLogTableInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogTableInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.logService.GetLogTableInfo(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetLogTableInfoError,
				code.Text(code.GetLogTableInfoError)+err.Error()).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
