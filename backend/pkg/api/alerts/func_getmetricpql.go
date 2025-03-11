// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetMetricPQL get metrics and PQL in alarm rules
// @Summary get metrics and PQL in alarm rules
// @Description get metrics and PQL in alarm rules
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetMetricPQLResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/metrics [get]
func (h *handler) GetMetricPQL() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetMetricPQLRequest)
		_ = c.ShouldBindQuery(&req)
		resp, err := h.alertService.GetMetricPQL(req)
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
