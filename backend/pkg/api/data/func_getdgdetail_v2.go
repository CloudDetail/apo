package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

type DGDetailReq struct {
	GroupID int64 `json:"groupId"`
}

func (h *handler) GetDGDetailV2() core.HandlerFunc {
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

		resp, err := h.dataService.GetGroupDetailWithSubGroup(c, req.GroupID)
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
