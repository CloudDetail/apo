package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AddAlertManagerConfigReceiver 新增告警通知对象
// @Summary 新增告警通知对象
// @Description 新增告警通知对象
// @Tags API.alerts
// @Accept json
// @Produce json
// @Param Request body request.AddAlertManagerConfigReceiver true "请求信息"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alerts/alertmanager/receiver/add [post]
func (h *handler) AddAlertManagerConfigReceiver() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AddAlertManagerConfigReceiver)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.AddAMConfigReceiver(req)
		if err != nil {

		}
		c.Payload("ok")
	}
}
