package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (h *handler) CreateDataGroupV2() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CreateDataGroupRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataService.CreateDataGroupV2(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CreateDataGroupError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
