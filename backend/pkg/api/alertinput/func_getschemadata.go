// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// GetSchemaData core.HandlerFunc
// @Summary core.HandlerFunc
// @Description core.HandlerFunc
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.AlertSchemaRequest true "Schema Info"
// @Success 200 {object} alert.GetSchemaDataReponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/schema/data/get [get]
func (h *handler) GetSchemaData() core.HandlerFunc {
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

		columns, rows, err := h.inputService.GetSchemaData(req.Schema)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetSchemaDataFailed,
				err,
			)
			return
		}
		c.Payload(alert.GetSchemaDataReponse{
			Columns: columns,
			Rows:    rows,
		})
	}
}
