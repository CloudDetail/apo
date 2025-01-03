// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetGroupList 获取group和label对应接口
// @Summary 获取group和label对应接口
// @Description 获取group和label对应接口
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetGroupListResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/groups [get]
func (h *handler) GetGroupList() core.HandlerFunc {
	return func(c core.Context) {
		resp := h.alertService.GetGroupList()
		c.Payload(resp)
	}
}
