// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// GetCluster GetCluster
// @Summary
// @Description
// @Tags API.integration
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body integration.Cluster true "请求信息"
// @Success 200 {object} response.getClusterResponse
// @Failure 400 {object} code.Failure
// @Router /api/integration/cluster/get [get]
func (h *handler) GetCluster() core.HandlerFunc {
	return func(c core.Context) {
		req := new(integration.Cluster)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		clusterIntegration, err := h.integrationService.GetClusterIntegration(req.ID)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetClusterIntegrationFailed,
				code.Text(code.GetClusterIntegrationFailed)).WithError(err),
			)
			return
		}
		c.Payload(clusterIntegration)
	}
}
