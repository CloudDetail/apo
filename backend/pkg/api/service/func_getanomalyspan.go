package service

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"
)

// GetAnomalySpan 获取服务和根因类型的故障报告
// @Summary 获取服务和根因类型的故障报告
// @Description 获取服务和根因类型的故障报告
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetAnomalySpanRequest true "请求信息"
// @Success 200 {object} response.GetAnomalySpanResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/anomaly-span/list [post]
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

		resp, err := h.serviceInfoService.GetAnomalySpan(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAnomalySpanError,
				code.Text(code.GetAnomalySpanError)))
			return
		}
		c.Payload(resp)
	}
}
