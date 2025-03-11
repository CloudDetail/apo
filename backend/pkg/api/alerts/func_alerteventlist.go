// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// AlertEventList
// @Summary
// @Description
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// TODO 下面的请求参数类型和返回类型需根据实际需求进行变更
// @Param Request body request.alertEventListRequest true "请求信息"
// @Success 200 {object} response.alertEventListResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/events/list [post]
func (h *handler) AlertEventList() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.AlertEventSearchRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		list, err := h.alertService.AlertEventList(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.GetAlertEventListError,
				c.ErrMessage(code.GetAlertEventListError)).WithError(err),
			)
			return
		}
		c.Payload(list)
	}
}
