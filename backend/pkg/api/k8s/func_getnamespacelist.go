// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetNamespaceList get all namespace information
// @Summary get all namespace information
// @Description get all namespace information
// @Tags API.k8s
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/k8s/namespaces [get]
func (h *handler) GetNamespaceList() core.HandlerFunc {
	return func(c core.Context) {

		resp, err := h.k8sService.GetNamespaceList()
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
