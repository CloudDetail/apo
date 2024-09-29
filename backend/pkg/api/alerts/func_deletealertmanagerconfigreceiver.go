package alerts

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteAlertManagerConfigReceiver 删除告警通知对象
// @Summary 删除告警通知对象
// @Description 删除告警通知对象
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.DeleteAlertManagerConfigReceiverRequest true "删除对象"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alertmanager/receiver [delete]
func (h *handler) DeleteAlertManagerConfigReceiver() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteAlertManagerConfigReceiverRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.DeleteAMConfigReceiver(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					code.Text(vErr.Code),
				).WithError(err))
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.DeleteAlertRuleError,
					code.Text(code.DeleteAlertRuleError),
				).WithError(err))
			}
			return
		}

		c.Payload("ok")
	}
}
