// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// ListCluster ListCluster
// @Summary ListCluster
// @Description ListCluster
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} integration.ListClusterResponse
// @Failure 400 {object} code.Failure
// @Router /api/integration/cluster/list [get]
func (h *handler) ListCluster() core.HandlerFunc {
	return func(c core.Context) {
		clusters, err := h.integrationService.ListCluster()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ListClusterFailed,
				code.Text(code.ListClusterFailed)).WithError(err),
			)
			return
		}
		c.Payload(integration.ListClusterResponse{
			Clusters: clusters,
		})
	}
}
