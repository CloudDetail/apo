package serviceoverview

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetThreshold 获取单个阈值配置信息
// @Summary 获取单个阈值配置信息
// @Description 获取单个阈值配置信息
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param serviceName query string false "应用名称"
// @Param endpoint query string false "endpoint"
// @Param level query string true "阈值等级"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200
// @Failure 400 {object} code.Failure
// @Router /api/service/getThreshold [get]
func (h *handler) GetThreshold() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetThresholdRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		serviceName := req.ServiceName
		endPoint := req.Endpoint
		level := req.Level
		resp, err := h.serviceoverview.GetThreshold(level, serviceName, endPoint)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetThresholdError,
				code.Text(code.GetThresholdError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
