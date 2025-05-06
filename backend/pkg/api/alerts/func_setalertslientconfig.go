package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// SetAlertSlienceConfig
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.SetAlertSlienceConfigRequest true "请求信息"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/alerts/slient [post]
func (h *handler) SetAlertSlienceConfig() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetAlertSlienceConfigRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.SetSlienceConfigByAlertID(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.SetAlertSlienceError,
				c.ErrMessage(code.SetAlertSlienceError)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
