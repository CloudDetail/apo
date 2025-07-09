// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// QueryServiceName Get service name by instance.
// @Summary Get service name by instance.
// @Description Get service name by instance.
// @Tags API.dataplane
// @Accept json
// @Produce json
// @Param Request body request.QueryServiceNameRequest true "Request"
// @Success 200 {object} response.QueryServiceNameResponse
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/servicename [post]
func (h *handler) QueryServiceName() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.QueryServiceNameRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		c.Payload(h.dataplaneService.GetServiceName(c, req))
	}
}
