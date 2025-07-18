// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetPolarisInfer access to Polaris metric analysis
// @Summary Get Polaris metric Analysis
// @Description Get Polaris metric Analysis
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param step query int64 true "query step (us)"
// @Param service query string true "Query service name"
// @Param endpoint query string true "Query Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetPolarisInferResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/polaris/infer [post]
func (h *handler) GetPolarisInfer() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetPolarisInferRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allowed, err := h.dataService.CheckServicesPermission(c, req.Service); !allowed || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, nil)
			return
		}

		res, err := h.serviceInfoService.GetPolarisInfer(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetPolarisInferError,
				err,
			)
			return
		}

		c.Payload(res)
	}
}
