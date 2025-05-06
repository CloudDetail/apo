package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

type MarkAlertResolvedManuallyRequest struct {
	AlertID string `json:"alertId" form:"alertId"`
}

// MarkAlertResolvedManually
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.markAlertResolvedManuallyRequest true "请求信息"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/alerts/resolve [post]
func (h *handler) MarkAlertResolvedManually() core.HandlerFunc {
	return func(c core.Context) {
		req := new(MarkAlertResolvedManuallyRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.alertService.ManualResolveLatestAlertEventByAlertID(req.AlertID)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertSlienceError,
				c.ErrMessage(code.MarkAlertResolvedError)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
