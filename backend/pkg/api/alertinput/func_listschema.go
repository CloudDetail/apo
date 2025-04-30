// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// ListSchema ListSchema
// @Summary ListSchema
// @Description ListSchema
// @Tags API.ListSchema
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} alert.ListSchemaResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/schema/list [get]
func (h *handler) ListSchema() core.HandlerFunc {
	return func(c core.Context) {

		schemas, err := h.inputService.ListSchema()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ListSchemaFailed,
				c.ErrMessage(code.ListSchemaFailed)).WithError(err),
			)
			return
		}

		c.Payload(alert.ListSchemaResponse{
			Schemas: schemas,
		})
	}
}
