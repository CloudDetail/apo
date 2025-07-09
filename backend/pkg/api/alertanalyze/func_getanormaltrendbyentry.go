package alertanalyze

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAnormalTrendByEntry 获取下游异常变化趋势
// @Summary 获取下游异常变化趋势
// @Description 获取下游异常变化趋势
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetDescendantAnormalTrendRequest true "请求信息"
// @Success 200 {object} response.GetDescendantDeltaAnormalEventResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/descendant/anormal/trend [post]
func (h *handler) GetAnormalTrendByEntry() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDescendantAnormalTrendRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.alertanalyzeService.GetAnormalTrendByEntry(c, req)
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
