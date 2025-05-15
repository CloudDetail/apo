// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// OtherTableInfo get external log table information
// @Summary get external log table information
// @Description get external log table information
// @Tags API.log
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.OtherTableInfoRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.OtherTableInfoResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/other/table [get]
func (h *handler) OtherTableInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.OtherTableInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp, err := h.logService.OtherTableInfo(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetOtherLogTableError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
