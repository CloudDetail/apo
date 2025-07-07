// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetErrorInstance get error instance
// @Summary get the error instance
// @Description get the error instance
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param startTime query uint64 true "query start time"
// @Param endTime query uint64 true "query end time"
// @Param service query string true "Query service name"
// @Param endpoint query string true "Query Endpoint"
// @Param step query int64 true "query step (us)"
// @Param entryService query string false "Ingress service name"
// @Param entryEndpoint query string false "entry Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetErrorInstanceResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/error/instance [post]
func (h *handler) GetErrorInstance() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetErrorInstanceRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allowed, err := h.dataService.CheckScopePermission(c, "", "", req.Service); !allowed || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, &response.GetErrorInstanceResponse{
				Status:    model.STATUS_NORMAL,
				Instances: []*response.ErrorInstance{},
			})
			return
		}

		resp, err := h.serviceInfoService.GetErrorInstance(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetErrorInstanceError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
