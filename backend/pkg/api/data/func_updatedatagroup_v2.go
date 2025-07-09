// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (h *handler) UpdateDataGroupV2() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateDataGroupRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataService.UpdateDataGroupV2(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UpdateDataGroupError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
