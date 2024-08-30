package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

// GetAlertEventsSample 获取采样告警事件
// @Summary 获取采样告警事件
// @Description 获取采样告警事件
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param group query string false "查询告警类型"
// @Success 200 {object} response.GetAlertEventsSampleResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/alert/sample/events [get]
func (h *handler) GetAlertEventsSample() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertEventsSampleRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.GetAlertEventsSample(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertEventsError,
				code.Text(code.GetAlertEventsError)).WithError(err),
			)
			return
		}
		if resp == nil {
			resp = &response.GetAlertEventsSampleResponse{
				EventMap: map[string]map[string][]clickhouse.AlertEventSample{},
			}
		}
		c.Payload(resp)
	}
}
