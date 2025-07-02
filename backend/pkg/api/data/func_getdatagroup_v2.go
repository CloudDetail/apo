package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

func (h *handler) GetDataGroupV2() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.dataService.ListDataGroupV2(c)
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
