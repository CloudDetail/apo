// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
)

// TriggerAdapterUpdate
// @Summary
// @Description
// @Tags API.integration
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/integration/adapter/update [get]
func (h *handler) TriggerAdapterUpdate() core.HandlerFunc {
	return func(c core.Context) {
		req := new(integration.TriggerAdapterUpdateRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		h.integrationService.TriggerAdapterUpdate(c, req)
		c.Payload("ok")
	}
}
