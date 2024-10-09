package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (h *handler) GetLogChart() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogQueryRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if req.Query == "" {
			req.Query = "(1='1')"
		}
		resp, err := h.logService.GetLogChart(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetLogChartError,
				code.Text(code.GetLogChartError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
