// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetServiceEndpointTopology get the upstream and downstream topology of a service
// @Summary get the upstream and downstream topology of the service
// @Description get the upstream and downstream topology of the service
// @Tags API.service
// @Accept application/json
// @Produce json
// @Param Request body request.GetServiceEndpointTopologyRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetServiceEndpointTopologyResponse
// @Failure 400 {object} code.Failure
// @Router /api/service/topology [post]
func (h *handler) GetServiceEndpointTopology() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetServiceEndpointTopologyRequest)
		if err := c.ShouldBind(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if allow, err := h.dataService.CheckGroupPermission(c, req.GroupID); !allow || err != nil {
			c.AbortWithPermissionError(err, code.AuthError, &response.GetServiceEndpointTopologyResponse{
				Parents:  []*model.TopologyNode{},
				Current:  &model.TopologyNode{},
				Children: []*model.TopologyNode{},
			})
			return
		}

		resp, err := h.serviceInfoService.GetServiceEndpointTopology(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetServiceUrlTopologyError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
