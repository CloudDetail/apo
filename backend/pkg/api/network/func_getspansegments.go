package deepflow

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"
)

// GetSpanSegmentsMetrics 客户端对外调用Span网络耗时分段指标
// @Summary 客户端对外调用Span网络耗时分段指标
// @Description 客户端对外调用Span网络耗时分段指标
// @Tags API.Network
// @Accept application/x-www-form-urlencoded
// @Param traceId query string true "traceId"
// @Param spanId query string false "spanId, 值为空则查询所有"
// @Success 200 {object} response.SpanSegmentMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/network/segments [get]
func (h *handler) GetSpanSegmentsMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SpanSegmentMetricsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.networkService.GetSpanSegmentsMetrics(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
