package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAlertRule 获取基础告警规则
// @Summary 获取基础告警规则
// @Description 获取基础告警规则
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param alertRuleFile query string false "查询告警规则文件名,为空返回所有"
// @Success 200 {object} response.GetAlertRuleResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rules [get]
func (h *handler) GetAlertRule() core.HandlerFunc {
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

		resp, err := h.alertService.GetAlertRule(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertRuleError,
				code.Text(code.GetAlertRuleError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
