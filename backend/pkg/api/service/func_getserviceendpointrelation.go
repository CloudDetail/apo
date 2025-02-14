// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetServiceEndpointRelation get the call relationship between the upstream and downstream services.
// @Summary get the call relationship between the upstream and downstream services
// @Description the call relationship between the upstream and downstream service
// @Tags API.service
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param startTime query uint64 true "query start time"
// @Param endTime query uint64 true "query end time"
// @Param service query string true "Query service name"
// @Param endpoint query string true "Query Endpoint"
// @Param entryService query string false "Ingress service name"
// @Param entryEndpoint query string false "entry Endpoint"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetServiceEndpointRelationResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/relation [get]
func (h *handler) GetServiceEndpointRelation() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceEndpointRelationRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, 0, nil, &req.Service, model.DATASOURCE_CATEGORY_APM)
		if err != nil {
			c.HandleError(err, code.AuthError, &response.GetServiceEndpointRelationResponse{
				Parents:       []*model.TopologyNode{},
				Current:       &model.TopologyNode{},
				ChildRelation: []*model.ToplogyRelation{},
			})
			return
		}
		resp, err := h.serviceInfoService.GetServiceEndpointRelation(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetServiceUrlRelationError,
				code.Text(code.GetServiceUrlRelationError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
