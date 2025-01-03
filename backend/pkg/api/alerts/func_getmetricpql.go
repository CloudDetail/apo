// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetMetricPQL 获取告警规则中指标和PQL
// @Summary 获取告警规则中指标和PQL
// @Description 获取告警规则中指标和PQL
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetMetricPQLResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/metrics [get]
func (h *handler) GetMetricPQL() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.alertService.GetMetricPQL()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetMetricPQLError,
				code.Text(code.GetMetricPQLError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
