// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	model "github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// CreateClusterIdentity registers a cluster without creating integration configuration.
// @Summary Create Cluster Identity
// @Description Register cluster identity without integration configuration
// @Tags API.integration
// @Accept application/json
// @Produce json
// @Param Request body model.CreateClusterIdentityRequest true "Cluster Identity"
// @Success 200 {object} model.Cluster "created cluster identity"
// @Failure 400 {object} code.Failure
// @Router /api/integration/cluster/identity [post]
func (h *handler) CreateClusterIdentity() core.HandlerFunc {
	return func(c core.Context) {
		req := new(model.CreateClusterIdentityRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}

		cluster, err := h.integrationService.CreateClusterIdentity(c, req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.CreateClusterFailed, err)
			return
		}
		c.Payload(cluster)
	}
}
