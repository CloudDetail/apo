// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// CreateCluster Create Cluster
// @Summary Create Cluster
// @Description Create Cluster
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body integration.Cluster true "Cluster Info"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/integration/cluster/create [post]
func (h *handler) CreateCluster() core.HandlerFunc {
	return func(c core.Context) {
		req := new(integration.ClusterIntegrationVO)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.integrationService.CreateCluster(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.CreateClusterFailed,
				code.Text(code.CreateClusterFailed)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
