package alerts

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UpdateAlertManagerConfigReceiver 更新告警通知对象
// @Summary 更新告警通知对象
// @Description 更新告警通知对象
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.UpdateAlertManagerConfigReceiver true "请求信息"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alertmanager/receiver [post]
func (h *handler) UpdateAlertManagerConfigReceiver() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateAlertManagerConfigReceiver)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.UpdateAMConfigReceiver(req)
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
					code.UpdateAMConfigReceiverError,
					code.Text(code.UpdateAMConfigReceiverError),
				).WithError(err))
			}
			return
		}
		c.Payload("ok")
	}
}
