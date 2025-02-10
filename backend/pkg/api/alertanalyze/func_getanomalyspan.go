package alertanalyze

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAnomalySpan 获取服务和根因类型的故障报告
// @Summary 获取服务和根因类型的故障报告
// @Description 获取服务和根因类型的故障报告
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetAnomalySpanRequest true "请求信息"
// @Success 200 {object} response.GetAnomalySpanResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/anomaly-span/list [post]
func (h *handler) GetAnomalySpan() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAnomalySpanRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		if !model.CheckPolarisType(req.Reason) {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)),
			)
			return
		}

		resp, err := h.alertanalyzeService.GetAnomalySpan(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAnomalySpanError,
				code.Text(code.GetAnomalySpanError)).WithError(err))
			return
		}
		c.Payload(resp)
	}
}
