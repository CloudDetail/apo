package metric

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/services/metric"
)

// QueryMetrics
// @Summary
// @Description
// @Tags API.metric
// @Accept application/x-www-form-urlencoded
// @Produce json
// TODO 下面的请求参数类型和返回类型需根据实际需求进行变更
// @Param Request body request.queryMetricsRequest true "请求信息"
// @Success 200 {object} response.queryMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/metric/query [post]
func (h *handler) QueryMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(metric.QueryMetricsRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp := h.metricService.QueryMetrics(req)
		c.Payload(resp)
	}
}
