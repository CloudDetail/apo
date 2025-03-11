// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetGroupList get the corresponding interfaces of group and label
// @Summary get the corresponding interfaces of group and label
// @Description get the corresponding interfaces of group and label
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetGroupListResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/groups [get]
func (h *handler) GetGroupList() core.HandlerFunc {
	return func(c core.Context) {
		resp := h.alertService.GetGroupList(c)
		c.Payload(resp)
	}
}
