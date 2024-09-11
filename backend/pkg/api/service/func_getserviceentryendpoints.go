package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetServiceEntryEndpoints 获取服务入口Endpoint列表
// @Summary 获取服务入口Endpoint列表
// @Description 获取服务入口Endpoint列表
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query uint64 true "查询开始时间"
// @Param endTime query uint64 true "查询结束时间"
// @Param service query string true "查询服务名"
// @Param endpoint query string true "查询Endpoint"
// @Success 200 {object} response.GetServiceEntryEndpointsResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/entry/endpoints [get]
func (h *handler) GetServiceEntryEndpoints() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceEntryEndpointsRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.GetServiceEntryEndpoints(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceEntryEndpointsError,
				code.Text(code.GetServiceEntryEndpointsError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
