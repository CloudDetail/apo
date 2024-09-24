package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetMetricPQL 获取告警规则中指标和PQL
// @Summary 获取告警规则中指标和PQL
// @Description 获取告警规则中指标和PQL
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} response.GetMetricPQLResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/metrics [get]
func (h *handler) GetMetricPQL() core.HandlerFunc {
	return func(c core.Context) {
		resp := h.alertService.GetMetricPQL()

		c.Payload(resp)
	}
}
