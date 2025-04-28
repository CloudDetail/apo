// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// DeleteSchema DeleteSchema
// @Summary DeleteSchema
// @Description DeleteSchema
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.AlertSchemaRequest true "SchemaInfo"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/schema/delete [post]
func (h *handler) DeleteSchema() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.AlertSchemaRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.inputService.DeleteSchema(c, req.Schema)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.DeleteSchemaFailed,
				c.ErrMessage(code.DeleteSchemaFailed)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
