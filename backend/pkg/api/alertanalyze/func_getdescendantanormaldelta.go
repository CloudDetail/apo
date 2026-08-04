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
// @Param Request body request.GetDescendantAnormalDeltaEventRequest true "请求信息"
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
