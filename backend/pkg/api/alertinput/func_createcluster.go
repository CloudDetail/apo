// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

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
// @Router /api/alertinput/cluster/create [post]
func (h *handler) CreateCluster() core.HandlerFunc {
	return func(c core.Context) {
		req := new(integration.Cluster)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.inputService.CreateCluster(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CreateClusterFailed,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
