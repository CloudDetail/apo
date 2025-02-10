package alertanalyze

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetPredefinedDetectExpr 获取预定义检测表达式
// @Summary 获取预定义检测表达式
// @Description 获取预定义检测表达式
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} response.GetPredefinedDetectExprResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/detect/mutation/metrics [get]
func (h *handler) GetPredefinedDetectExpr() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.alertanalyzeService.DetectDefectsOptions()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetQuickMutationMetricError,
				code.Text(code.GetQuickMutationMetricError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
