// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// GetIntegrationInstallDoc
// @Summary
// @Description
// @Tags API.integration
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body integration.GetCInstallRequest true "请求信息"
// @Success 200 {object} integration.GetCInstallRequest
// @Failure 400 {object} code.Failure
// @Router /api/integration/cluster/install/cmd [get]
func (h *handler) GetIntegrationInstallDoc() core.HandlerFunc {
	return func(c core.Context) {
		req := new(integration.GetCInstallRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.integrationService.GetIntegrationInstallDoc(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetIntegrationInstallDocFailed,
				err,
			)
			return
		}
		c.Payload(integration.GetCInstallDocResponse{
			InstallMD: resp,
		})
	}
}
