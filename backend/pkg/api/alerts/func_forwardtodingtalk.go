// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// ForwardToDingTalk the received alarm is forwarded to the DingTalk
// @Summary the received alarm is forwarded to the DingTalk
// @Description the received alarm is forwarded to the DingTalk
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.ForwardToDingTalkRequest true "Request information"
// @Param uuid path string true "DingTalk the uuid corresponding to the webhook"
// @Success 200
// @Failure 400 {object} code.Failure
// @Router /api/alerts/outputs/dingtalk/{uuid} [post]
func (h *handler) ForwardToDingTalk() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.ForwardToDingTalkRequest)
		uuid := c.Param("uuid")
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		if err := h.alertService.ForwardToDingTalk(req, uuid); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest, "", ""))
		}
		c.Payload("OK")
	}
}
