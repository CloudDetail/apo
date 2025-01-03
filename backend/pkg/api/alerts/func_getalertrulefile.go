// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAlertRuleFile 获取基础告警规则
// @Summary 获取基础告警规则
// @Description 获取基础告警规则
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param alertRuleFile query string false "查询告警规则文件名,为空返回所有"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetAlertRuleFileResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rules [get]
func (h *handler) GetAlertRuleFile() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertRuleConfigRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.alertService.GetAlertRuleFile(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertRuleError,
				code.Text(code.GetAlertRuleError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
