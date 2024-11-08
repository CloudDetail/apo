package k8s

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"
)

// GetPodList 获取namespace下所有pod信息
// @Summary 获取所有pod信息
// @Description 获取所有pod信息
// @Tags API.k8s
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param namespace query string true "namespace名"
// @Success 200 {object} response.GetPodListResponse
// @Failure 400 {object} code.Failure
// @Router /api/k8s/pods [get]
func (h *handler) GetPodList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetPodListRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.k8sService.GetPodList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.K8sGetResourceError,
				code.Text(code.K8sGetResourceError)).WithError(err))
			return
		}

		c.Payload(resp)
	}
}
