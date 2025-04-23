package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AlertEventDetail
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetAlertDetailRequest true "请求信息"
// @Success 200 {object} response.GetAlertDetailResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/events/detail [post]
func (h *handler) AlertEventDetail() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertDetailRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.alertService.AlertDetail(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertEventListError,
				c.ErrMessage(code.GetAlertEventListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
