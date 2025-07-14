// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (h *handler) GetFilterByGroupIDV2() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DGFilterRequest)
		err := c.ShouldBindJSON(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp, err := h.dataService.GetFilterByGroupID(c, req)
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
