// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package integration

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// GetIntegrationVar GetIntegrationVar
// @Summary
// @Description
// @Tags API.integration
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body integration.GetIntegrationVarRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/integration/vars/:variable [get]
func (h *handler) GetIntegrationVar() core.HandlerFunc {
	return func(c core.Context) {
		var req = new(integration.GetIntegrationVarRequest)
		if err := c.ShouldBindURI(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		switch strings.ToLower(req.Variable) {
		case "baseurl":
			c.Payload(h.integrationService.GetBaseURL())
		case "serveraddr":
			c.Payload(h.integrationService.GetServerAddr())
		default:
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				fmt.Errorf("variable %s not found", req.Variable),
			)
			return
		}
	}
}
