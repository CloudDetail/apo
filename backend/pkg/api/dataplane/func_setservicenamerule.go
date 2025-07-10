// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// SetServiceNameRule Create or update servicename rule.
// @Summary Create or update servicename rule.
// @Description Create or update servicename rule.
// @Tags API.dataplane
// @Accept json
// @Produce json
// @Param Request body request.SetServiceNameRuleRequest true "Request"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/servicename/upsertRule [post]
func (h *handler) SetServiceNameRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetServiceNameRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		err := h.dataplaneService.SetServiceNameRule(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.SetServiceNameRuleError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
