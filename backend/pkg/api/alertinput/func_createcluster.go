// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// CreateCluster 创建集群
// @Summary 创建集群
// @Description 创建集群
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.Cluster true "请求信息"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/cluster/create [post]
func (h *handler) CreateCluster() core.HandlerFunc {
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

		err := h.inputService.CreateCluster(req)
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
