package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

func (h *handler) GetDGScopeList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(DGDetailReq)
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp, err := h.dataService.ListDataScopeByGroupID(c, req.GroupID)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetDatasourceError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
