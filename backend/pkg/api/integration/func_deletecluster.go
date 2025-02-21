// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// DeleteCluster DeleteCluster
// @Summary DeleteCluster
// @Description DeleteCluster
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body integration.Cluster true "Cluster Info"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/integration/cluster/delete [get]
func (h *handler) DeleteCluster() core.HandlerFunc {
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

		err := h.integrationService.DeleteCluster(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.DeleteClusterFailed,
				code.Text(code.DeleteClusterFailed)).WithError(err),
			)
			return

		}
		c.Payload("ok")
	}
}
