package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAlertRules 列出告警规则
// @Summary 列出告警规则
// @Description 列出告警规则
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param alertRuleFile query string false "告警规则文件名,为空返回默认告警文件"
// @Param Request body request.GetAlertRuleRequest true "请求信息"
// @Success 200 {object} response.GetAlertRulesResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rules [get]
func (h *handler) GetAlertRules() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertRuleRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp := h.alertService.GetAlertRules(req.AlertRuleFile)
		c.Payload(resp)
	}
}
