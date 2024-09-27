package alerts

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// CheckAlertRule 检查告警规则名是否可用
// @Summary 检查告警规则名是否可用
// @Description 检查告警规则名是否可用
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param alertRuleFile query string false "查询告警规则文件名"
// @Param group query string true "组名"
// @Param alert query string true "告警规则名"
// @Success 200 {object} response.CheckAlertRuleResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/available  [get]
func (h *handler) CheckAlertRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CheckAlertRuleRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.alertService.CheckAlertRule(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code),
				).WithError(err),
				)
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AddAlertRuleError,
					code.Text(code.UpdateAlertRuleError),
				).WithError(err),
				)
			}
			return
		}

		c.Payload(resp)
	}
}
