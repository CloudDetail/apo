package serviceoverview

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetMonitorStatus 获取kuma监控的服务状态
// @Summary 获取kuma监控的服务状态
// @Description 获取kuma监控的服务状态
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetMonitorStatusResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/monitor/status [get]
func (h *handler) GetMonitorStatus() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetMonitorStatusRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		startTime := time.UnixMicro(req.StartTime)
		endTime := time.UnixMicro(req.EndTime)

		resp, err := h.serviceoverview.GetMonitorStatus(startTime, endTime)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetMonitorStatusError,
				code.Text(code.GetMonitorStatusError)))
		}
		c.Payload(resp)
	}
}
