// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// CreateSchema 创建映射结构
// @Summary 创建映射结构
// @Description 创建映射结构
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.CreateSchemaRequest true "请求信息"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/schema/create [post]
func (h *handler) CreateSchema() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.CreateSchemaRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.CreateSchemaFailed,
				code.Text(code.CreateSchemaFailed)).WithError(err),
			)
			return
		}

		err := h.inputService.CreateSchema(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.CreateSchemaFailed,
				code.Text(code.CreateSchemaFailed)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
