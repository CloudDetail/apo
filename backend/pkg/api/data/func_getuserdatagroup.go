// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package data

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetUserDataGroup Get user's assigned data group.
// @Summary Get user's assigned data group.
// @Description Get user's assigned data group.
// @Tags API.data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param userId query int64 true "user's id"
// @Param category query string false "apm or normal, return all if is empty"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetSubjectDataGroupResponse
// @Failure 400 {object} code.Failure
// @Router /api/data/user/group [get]
func (h *handler) GetUserDataGroup() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetUserDataGroupRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		subReq := new(request.GetSubjectDataGroupRequest)
		subReq.SubjectID = req.UserID
		subReq.SubjectType = model.DATA_GROUP_SUB_TYP_USER
		subReq.Category = req.Category
		groups, err := h.dataService.GetSubjectDataGroup(subReq)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetDataGroupError,
				err,
			)
			return
		}
		resp := make(response.GetSubjectDataGroupResponse, 0, len(groups))
		seen := make(map[int64]struct{})
		for _, group := range groups {
			if _, ok := seen[group.GroupID]; ok {
				continue
			}
			seen[group.GroupID] = struct{}{}
			resp = append(resp, group)
		}
		c.Payload(resp)
	}
}
