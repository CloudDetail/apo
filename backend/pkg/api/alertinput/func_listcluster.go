// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// ListCluster ListCluster
// @Summary ListCluster
// @Description ListCluster
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} alert.ListClusterResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/cluster/list [get]
func (h *handler) ListCluster() core.HandlerFunc {
	return func(c core.Context) {
		clusters, err := h.inputService.ListCluster()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ListClusterFailed,
				code.Text(code.ListClusterFailed)).WithError(err),
			)
			return
		}
		c.Payload(alert.ListClusterResponse{
			Clusters: clusters,
		})
	}
}