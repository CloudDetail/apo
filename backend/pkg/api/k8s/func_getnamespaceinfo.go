// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetNamespaceInfo get namespace information
// @Summary get namespace information
// @Description get namespace information
// @Tags API.k8s
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param namespace query string true "namespace name"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/k8s/namespace/info [get]
func (h *handler) GetNamespaceInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetNamespaceInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.k8sService.GetNamespaceInfo(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.K8sGetResourceError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
