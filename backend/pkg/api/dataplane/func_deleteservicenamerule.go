// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package dataplane

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// DeleteServiceNameRule Delete servicename rule.
// @Summary Delete servicename rule.
// @Description Delete servicename rule.
// @Tags API.dataplane
// @Accept json
// @Produce json
// @Param Request body request.DeleteServiceNameRuleRequest true "Delete ServiceName Rule Request"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/dataplane/servicename/deleteRule [post]
func (h *handler) DeleteServiceNameRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.DeleteServiceNameRuleRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.dataplaneService.DeleteServiceNameRule(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.DeleteServiceNameRuleError,
				err,
			)
			return
		}
		c.Payload("ok")
	}
}
