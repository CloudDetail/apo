// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetDescendantRelevance get the dependent node delay correlation degree
// @Summary get the dependent node delay correlation degree
// @Description get the dependent node delay correlation degree
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query int64 true "query start time"
// @Param endTime query int64 true "query end time"
// @Param service query string true "Query service name"
// @Param endpoint query string true "Query Endpoint"
// @Param step query int64 true "query step (us)"
// @Param entryService query string false "Ingress service name"
// @Param entryEndpoint query string false "entry Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} []response.GetDescendantRelevanceResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/descendant/relevance [get]
func (h *handler) GetDescendantRelevance() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDescendantRelevanceRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		res, err := h.serviceInfoService.GetDescendantRelevance(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetDescendantRelevanceError,
				code.Text(code.GetDescendantRelevanceError)).WithError(err),
			)
			return
		}
		c.Payload(res)
	}
}
