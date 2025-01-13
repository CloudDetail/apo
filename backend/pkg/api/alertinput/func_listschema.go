// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// ListSchema 列出映射结构
// @Summary 列出映射结构
// @Description 列出映射结构
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
				code.Text(code.ListSchemaFailed)).WithError(err),
			)
			return
		}

		c.Payload(alert.ListSchemaResponse{
			Schemas: schemas,
		})
	}
}
