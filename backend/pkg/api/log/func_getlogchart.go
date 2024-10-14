package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetLogChart 获取日志趋势图
// @Summary 获取日志趋势图
// @Description 获取日志趋势图
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.LogQueryRequest true "请求信息"
// @Success 200 {object} response.LogChartResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/chart [post]
func (h *handler) GetLogChart() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogQueryRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if req.Query == "" {
			req.Query = "(1='1')"
		}
		if req.TimeField == "" {
			req.TimeField = "timestamp"
		}
		if req.LogField == "" {
			req.LogField = "content"
		}
		resp, err := h.logService.GetLogChart(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetLogChartError,
				code.Text(code.GetLogChartError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
