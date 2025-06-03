// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetFaultLogContent get the contents of the fault site log
// @Summary get the contents of the fault site log
// @Description get the contents of the fault site log
// @Tags API.log
// @Produce json
// @Param Request body request.GetFaultLogContentRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetFaultLogContentResponse
// @Failure 400 {object} code.Failure
// @Router /api/log/fault/content [post]
func (h *handler) GetFaultLogContent() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetFaultLogContentRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		// TODO GetFaultLogContentRequest's service is unused, won't check data permission
		//userID := c.UserID()
		//err := h.dataService.CheckDatasourcePermission(c,userID, 0, nil, &req.ServiceName, "")
		//if err != nil {
		//	c.AbortWithPermissionError(err, code.AuthError)
		//	return
		//}

		resp, err := h.logService.GetFaultLogContent(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetFaultLogContentError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
