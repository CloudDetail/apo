// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"fmt"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// GetIntegrationInstallConfigFile
// @Summary
// @Description
// @Tags API.integration
// @Accept application/x-www-form-urlencoded
// @Produce text/plain
// @Param Request body integration.GetIntegrationInstallRequest true "请求信息"
// @Success 200 {object} response.getIntegrationInstallConfigFileResponse
// @Failure 400 {object} code.Failure
// @Router /api/integration/cluster/install/config [get]
func (h *handler) GetIntegrationInstallConfigFile() core.HandlerFunc {
	return func(c core.Context) {
		req := new(integration.GetCInstallRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		configFile, err := h.integrationService.GetIntegrationInstallConfigFile(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetIntegrationInstallConfigFileFailed,
				code.Text(code.GetIntegrationInstallConfigFileFailed)).WithError(err),
			)
			return
		}
		c.SetHeader("Access-Control-Expose-Headers", "Content-Disposition,Content-Type")
		c.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", configFile.FileName))
		c.SetHeader("Content-Type", "text/plain")
		c.Payload(configFile.Content)
	}
}
