package alertanalyze

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"

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
// @Router /api/alerts/descendant/anormal/trend [get]
func (h *handler) GetAnormalTrendByEntry() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDescendantAnormalTrendRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.alertanalyzeService.GetAnormalTrendByEntry(req)
		if err != nil {
			var vErr model.ErrMutationCheckFailed
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.MutationPQLCheckFailed,
					code.Text(code.MutationPQLCheckFailed),
				).WithError(err))
				return
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AlertAnalyzeDescendantAnormalEventDeltaError,
					code.Text(code.AlertAnalyzeDescendantAnormalEventDeltaError),
				).WithError(err))
				return
			}
		}
		c.Payload(resp)
	}
}
