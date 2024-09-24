package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteAlertManagerConfigReceiver 删除告警通知对象
// @Summary 删除告警通知对象
// @Description 删除告警通知对象
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param amConfigFile query string false "告警通知配置文件名"
// @Param name query string false "告警通知配置名称"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alertmanager/receiver [delete]
func (h *handler) DeleteAlertManagerConfigReceiver() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteAlertManagerConfigReceiverRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp := h.alertService.DeleteAMConfigReceiver(req)
		c.Payload(resp)
	}
}
