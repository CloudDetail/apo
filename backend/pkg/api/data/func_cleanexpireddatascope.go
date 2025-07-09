// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (h *handler) CleanExpiredDataScope() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CleanExpiredDataScopeRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.dataService.CleanExpiredDataScope(c, req.GroupID, req.Clean)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.CleanExpiredDataScopeError,
				err,
			)
			return
		}

		c.Payload(resp)
	}
}
