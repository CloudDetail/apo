// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetDescendantRelevance get the dependent node delay correlation degree
// @Summary get the dependent node delay correlation degree
// @Description get the dependent node delay correlation degree
// @Tags API.service
// @Accept application/json
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
// @Router /api/service/descendant/relevance [post]
func (h *handler) GetDescendantRelevance() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetDescendantRelevanceRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allow, err := h.dataService.CheckGroupPermission(c, req.GroupID); !allow || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, []response.GetDescendantRelevanceResponse{})
			return
		}

		res, err := h.serviceInfoService.GetDescendantRelevance(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetDescendantRelevanceError,
				err,
			)
			return
		}
		c.Payload(res)
	}
}
