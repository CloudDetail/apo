package deepflow

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"net/http"
)

func (h *handler) GetPodMap() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.PodMapRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		resp, err := h.networkService.GetPodMap(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ServerError,
				code.Text(code.ServerError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
