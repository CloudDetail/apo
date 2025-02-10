package alertanalyze

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetDescendantAnormal 获取下游告警事件
// @Summary 获取下游告警事件
// @Description 获取下游告警事件
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Param entryService query string false "入口服务名"
// @Param entryEndpoint query string false "入口Endpoint"
// @Success 200 {object} response.GetDescendantAnormalEventResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/descendant/anormal/list [get]
func (h *handler) GetDescendantAnormal() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDescendantAnormalEventRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.alertanalyzeService.GetDescendantAnormal(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.AlertAnalyzeDescendantAnormalEventError,
				code.Text(code.AlertAnalyzeDescendantAnormalEventError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
