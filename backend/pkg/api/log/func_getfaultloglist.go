// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetFaultLogPageList get the fault site paging log
// @Summary get fault site paging log
// @Description get fault site paging log
// @Tags API.log
// @Accept json
// @Produce json
// @Param Request body request.GetFaultLogPageListRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetFaultLogPageListResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/fault/pagelist [post]
func (h *handler) GetFaultLogPageList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetFaultLogPageListRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if req.PageNum == 0 {
			req.PageNum = 1
		}
		if req.PageSize == 0 {
			req.PageSize = 10
		}
		userID := middleware.GetContextUserID(c)
		err := h.dataService.CheckDatasourcePermission(userID, req.GroupID, &req.Namespaces, &req.Service, "")
		if err != nil {
			c.HandleError(err, code.AuthError, &response.GetFaultLogPageListResponse{
				Pagination: &model.Pagination{
					Total:       0,
					CurrentPage: 0,
					PageSize:    0,
				},
				List: []clickhouse.FaultLogResult{},
			})
			return
		}
		resp, err := h.logService.GetFaultLogPageList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetFaultLogPageListError,
				code.Text(code.GetFaultLogPageListError)).WithError(err),
			)
			return
		}
		c.Payload(resp)
	}
}
