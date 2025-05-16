// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// CheckSchemaIsUsed Check whether the mapping structure is used
// @Summary Check whether the mapping structure is used
// @Description Check whether the mapping structure is used
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.AlertSchemaRequest true "Schema Info"
// @Success 200 {object} alert.CheckSchemaIsUsedReponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/schema/used/check [get]
func (h *handler) CheckSchemaIsUsed() core.HandlerFunc {
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

		alertSources, err := h.inputService.CheckSchemaIsUsed(c, req.Schema)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CheckSchemaUsedFailed,
				err,
			)
			return
		}
		c.Payload(alert.CheckSchemaIsUsedReponse{
			IsUsing:          len(alertSources) > 0,
			AlertSourceNames: alertSources,
		})
	}
}
