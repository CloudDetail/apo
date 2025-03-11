// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// ListSchemaWithColumns ListSchemaWithColumns
// @Summary ListSchemaWithColumns
// @Description ListSchemaWithColumns
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} alert.ListSchemaWithColumnsResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/schema/listwithcolumns [get]
func (h *handler) ListSchemaWithColumns() core.HandlerFunc {
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

		var resp = alert.ListSchemaWithColumnsResponse{
			Schemas: make(map[string][]string, len(schemas)),
		}
		for _, schema := range schemas {
			columns, err := h.inputService.ListSchemaColumns(schema)
			if err != nil {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.ListSchemaFailed,
					c.ErrMessage(code.ListSchemaFailed)).WithError(err),
				)
				return
			}
			resp.Schemas[schema] = columns
		}
		c.Payload(resp)
	}
}
