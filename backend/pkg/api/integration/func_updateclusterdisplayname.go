// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// UpdateClusterDisplayName updates the display name without rewriting integration settings.
// @Summary Update cluster display name
// @Tags API.integration
// @Accept application/json
// @Produce json
// @Param Request body integration.UpdateClusterDisplayNameRequest true "Cluster display name"
// @Success 200 {object} integration.Cluster
// @Failure 400 {object} code.Failure
// @Router /api/integration/cluster/display-name [post]
func (h *handler) UpdateClusterDisplayName() core.HandlerFunc {
	return func(c core.Context) {
		req := new(integration.UpdateClusterDisplayNameRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		cluster, err := h.integrationService.UpdateClusterDisplayName(c, req.ID, req.Name)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.UpdateClusterFailed, err)
			return
		}
		c.Payload(cluster)
	}
}
