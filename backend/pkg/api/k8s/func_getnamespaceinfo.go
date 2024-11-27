package k8s

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetNamespaceInfo 获取namespace信息
// @Summary 获取namespace信息
// @Description 获取namespace信息
// @Tags API.k8s
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param namespace query string true "namespace名"
// @Success 200 {object} string
// @Failure 400 {object} code.Failure
// @Router /api/k8s/namespace/info [get]
func (h *handler) GetNamespaceInfo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetNamespaceInfoRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.k8sService.GetNamespaceInfo(req)
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
