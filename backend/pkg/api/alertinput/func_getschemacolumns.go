// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// GetSchemaColumns GetSchemaColumns
// @Summary GetSchemaColumns
// @Description GetSchemaColumns
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.AlertSchemaRequest true "Schema Info"
// @Success 200 {object} alert.GetSchemaColumnsResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/schema/column/get [get]
func (h *handler) GetSchemaColumns() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.AlertSchemaRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		columns, err := h.inputService.ListSchemaColumns(c, req.Schema)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetSchemaColumnsFailed,
				err,
			)
			return
		}
		c.Payload(alert.GetSchemaColumnsResponse{
			Columns: columns,
		})
	}
}
