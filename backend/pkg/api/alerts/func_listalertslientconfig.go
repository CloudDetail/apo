package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// ListAlertSlienceConfig
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.ListAlertSlienceConfig true "请求信息"
// @Success 200 {object} response.ListAlertSlienceConfigResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/slient/list [get]
func (h *handler) ListAlertSlienceConfig() core.HandlerFunc {
	return func(c core.Context) {
		sliences, err := h.alertService.ListSlienceConfig()
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertSlienceError,
				c.ErrMessage(code.ListAlertSlienceError)).WithError(err),
			)
			return
		}
		c.Payload(response.ListAlertSlienceConfigResponse{
			Sliences: sliences,
		})
	}
}
