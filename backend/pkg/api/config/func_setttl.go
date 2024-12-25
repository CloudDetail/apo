// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// SetTTL 配置TTL
// @Summary  配置TTL
// @Description  配置TTL
// @Tags Api.config
// @Accept json
// @Produce json
// @Param Request body request.SetTTLRequest true "请求信息"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200
// @Failure 400
// @Router /api/config/setTTL [post]
func (h *handler) SetTTL() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.SetTTLRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}
		if err := h.configService.SetTTL(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.SetTTLError,
				code.Text(code.SetTTLError)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
