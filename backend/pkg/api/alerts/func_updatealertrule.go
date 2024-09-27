package alerts

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UpdateAlertRule 更新告警规则
// @Summary 更新告警规则
// @Description 更新告警规则
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.UpdateAlertRuleRequest true "请求信息"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule [post]
func (h *handler) UpdateAlertRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateAlertRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.UpdateAlertRule(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code)).WithError(err),
				)
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.UpdateAlertRuleError,
					code.Text(code.UpdateAlertRuleError)).WithError(err),
				)
			}
			return
		}
		c.Payload("ok")
	}
}
