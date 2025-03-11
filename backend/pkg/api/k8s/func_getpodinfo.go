// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetPodInfo get pod information
// @Summary get pod information
// @Description get pod information
// @Tags API.k8s
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param namespace query string true "namespace name"
// @Param pod query string true "pod name"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/k8s/pod/info [get]
func (h *handler) GetPodInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetPodInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.k8sService.GetPodInfo(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.K8sGetResourceError,
				c.ErrMessage(code.K8sGetResourceError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
