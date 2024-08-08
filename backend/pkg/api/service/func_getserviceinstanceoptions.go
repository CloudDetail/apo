package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServiceInstanceOptions 获取服务下拉实例列表
// @Summary 获取服务下拉实例列表
// @Description 获取服务下拉实例列表
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "查询开始时间"
// @Param endTime query int64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Success 200 {object} map[string]model.ServiceInstance
// @Failure 400 {object} code.Failure
// @Router /api/service/instance/options [get]
func (h *handler) GetServiceInstanceOptions() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceInstanceOptionsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.GetServiceInstanceOptions(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceInstanceOptionsError,
				code.Text(code.GetServiceInstanceListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
