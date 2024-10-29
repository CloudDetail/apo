package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// ForwardToDingTalk 接收告警转发到钉钉
// @Summary 接收告警转发到钉钉
// @Description 接收告警转发到钉钉
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.ForwardToDingTalkRequest true "请求信息"
// @Param uuid path string true "钉钉webhook对应的uuid"
// @Success 200
// @Failure 400 {object} code.Failure
// @Router /api/alerts/outputs/dingtalk/{uuid} [post]
func (h *handler) ForwardToDingTalk() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ForwardToDingTalkRequest)
		uuid := c.Param("uuid")
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		if err := h.alertService.ForwardToDingTalk(req, uuid); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest, "", ""))
		}
		c.Payload("OK")
	}
}
