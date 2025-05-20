package alertanalyze

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetDescendantAnormalDelta 获取下游告警事件
// @Summary 获取下游告警事件
// @Description 获取下游告警事件
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param step query int64 true "查询步长(us)"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Param anormalTypes query string true "异常事件类型过滤"
// @Success 200 {object} response.GetDescendantAnormalEventResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/descendant/anormal/delta [post]
func (h *handler) GetDescendantAnormalDelta() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDescendantAnormalDeltaEventRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.alertanalyzeService.SearchAnormalDeltaByEntry(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.AlertAnalyzeDescendantAnormalEventDeltaError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
