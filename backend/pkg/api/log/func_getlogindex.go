package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (h *handler) GetLogIndex() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.LogIndexRequest)
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
		resp, err := h.logService.GetLogIndex(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetLogIndexError,
				code.Text(code.GetLogIndexError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
