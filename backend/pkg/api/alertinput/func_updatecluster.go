// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// UpdateCluster UpdateCluster
// @Summary UpdateCluster
// @Description UpdateCluster
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.Cluster true "Cluster Info"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/cluster/update [post]
func (h *handler) UpdateCluster() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.Cluster)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.inputService.UpdateCluster(req)
		if err != nil {

		}
		c.Payload("ok")
	}
}
