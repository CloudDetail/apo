package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetTraceMetrics 获取Trace相关指标
// @Summary 获取Trace相关指标
// @Description 获取Trace相关指标
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query uint64 true "查询开始时间"
// @Param endTime query uint64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Param step query int64 true "查询步长(us)"
// @Param entryService query string false "入口服务名"
// @Param entryEndpoint query string false "入口Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []response.GetTraceMetricsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/trace/metrics [get]
func (h *handler) GetTraceMetrics() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetTraceMetricsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.GetTraceMetrics(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetTraceMetricsError,
				code.Text(code.GetTraceMetricsError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
