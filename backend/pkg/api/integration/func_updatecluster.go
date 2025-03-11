// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// UpdateCluster UpdateCluster
// @Summary UpdateCluster
// @Description UpdateCluster
// @Tags API.integration
// @Accept application/json
// @Produce json
// @Param Request body integration.Cluster true "Cluster Info"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/integration/cluster/update [post]
func (h *handler) UpdateCluster() core.HandlerFunc {
	return func(c core.Context) {
		req := new(integration.ClusterIntegration)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.integrationService.UpdateClusterIntegration(req)
		if err != nil {

		}
		c.Payload("ok")
	}
}
