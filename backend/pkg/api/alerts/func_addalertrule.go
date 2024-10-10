package alerts

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// AddAlertRule 新增告警规则
// @Summary 新增告警规则
// @Description 新增告警规则
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.AddAlertRuleRequest true "请求信息"
// @Success 200
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/add [post]
func (h *handler) AddAlertRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AddAlertRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.AddAlertRule(req)
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
		c.Payload("ok")
	}
}
