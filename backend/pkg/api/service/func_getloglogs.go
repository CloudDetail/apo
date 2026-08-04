// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

// GetLogLogs get Log fault site log
// @Summary get Log fault site log
// @Description get Log fault site log
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param Request body request.GetLogLogsRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []clickhouse.FaultLogResult
// @Failure 400 {object} code.Failure
// @Router /api/service/log/logs [post]
func (h *handler) GetLogLogs() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetLogLogsRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allowed, err := h.dataService.CheckServicesPermission(c, req.Service); !allowed || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, []clickhouse.FaultLogResult{})
			return
		}

		resp, err := h.serviceInfoService.GetLogLogs(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetLogLogsError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
