package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// InputAlertManager 获取 AlertManager 的告警事件
// @Summary 获取 AlertManager 的告警事件
// @Description 获取 AlertManager 的告警事件
// @Tags API.alerts
// @Accept application/json
// @Produce json
// @Param Request body request.InputAlertManagerRequest true "请求信息"
// @Success 200 string ok
// @Failure 400 {object} code.Failure
// @Router /api/alerts/inputs/alertmanager [post]
func (h *handler) InputAlertManager() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.InputAlertManagerRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		if err := h.alertService.InputAlertManager(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.DbConnectError,
				code.Text(code.DbConnectError)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
