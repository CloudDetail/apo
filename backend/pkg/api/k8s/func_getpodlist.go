// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetPodList get information about all pods in the namespace
// @Summary get all pod information
// @Description get all pod information
// @Tags API.k8s
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param namespace query string true "namespace name"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/k8s/pods [get]
func (h *handler) GetPodList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetPodListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.k8sService.GetPodList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.K8sGetResourceError,
				code.Text(code.K8sGetResourceError)).WithError(err))
			return
		}

		c.Payload(resp)
	}
}
