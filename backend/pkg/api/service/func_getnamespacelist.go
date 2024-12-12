package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetNamespaceList Get monitored namespaces.
// @Summary Get monitored namespaces.
// @Description Get monitored namespaces.
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "开始时间"
// @Param endTime query int64 true "结束时间"
// @Success 200 {object} response.GetServiceNamespaceListResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/namespace/list [get]
func (h *handler) GetNamespaceList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceNamespaceListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.serviceInfoService.GetServiceNamespaceList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetNamespaceListError,
				code.Text(code.GetNamespaceListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
