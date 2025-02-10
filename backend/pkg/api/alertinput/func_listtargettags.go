// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

// GetTargetTags GetTargetTags
// @Summary GetTargetTags
// @Description GetTargetTags
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} alert.GetTargetTagsResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertsinput/enrich/tags/list [get]
func (h *handler) ListTargetTags() core.HandlerFunc {
	return func(c core.Context) {
		targetTags, err := h.inputService.GetAlertEnrichRuleTags()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertsInputTargetTagsError,
				code.Text(code.GetAlertsInputTargetTagsError)).WithError(err),
			)
			return
		}

		c.Payload(alert.GetTargetTagsResponse{
			TargetTags: targetTags,
		})
	}
}
