package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAlertManagerConfigReceiver 列出告警通知对象
// @Summary 列出告警通知对象
// @Description 列出告警通知对象
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.GetAlertManagerConfigReceverRequest true "请求信息"
// @Success 200 {object} response.GetAlertManagerConfigReceiverResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alertmanager/receiver/list [post]
func (h *handler) GetAlertManagerConfigReceiver() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertManagerConfigReceverRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp := h.alertService.GetAMConfigReceivers(req)
		c.Payload(resp)
	}
}
