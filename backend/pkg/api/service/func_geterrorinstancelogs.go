// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetErrorInstanceLogs get the error instance fault site log
// @Summary get the error instance failure site log
// @Description get the error instance failure site log
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param endpoint query string true "Query Endpoint"
// @Param instance query string false "instance name"
// @Param nodeName query string false "hostname"
// @Param containerId query string false "container name"
// @Param pid query uint32 false "process number"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []clickhouse.FaultLogResult
// @Failure 400 {object} code.Failure
// @Router /api/service/errorinstance/logs [get]
func (h *handler) GetErrorInstanceLogs() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetErrorInstanceLogsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		userID := c.UserID()
		err := h.dataService.CheckDatasourcePermission(c, userID, 0, nil, &req.Service, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.AbortWithPermissionError(err, code.AuthError, []clickhouse.FaultLogResult{})
			return
		}
		resp, err := h.serviceInfoService.GetErrorInstanceLogs(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetErrorInstanceLogsError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
