// Copyright 2024 CloudDetail
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
// @Produce string
// @Success 200 string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/integration/adapter/update [get]
func (h *handler) TriggerAdapterUpdate() core.HandlerFunc {
	return func(c core.Context) {
		req := new(integration.TriggerAdapterUpdateRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		h.integrationService.TriggerAdapterUpdate(req)
		c.Payload("ok")
	}
}
