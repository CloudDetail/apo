// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// UpdateSchemaData 更新映射结构中的数据
// @Summary 更新映射结构中的数据
// @Description 更新映射结构中的数据
// @Tags API.alertinput
// @Accept application/json
// @Produce json
// @Param Request body alert.UpdateSchemaDataRequest true "请求信息"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/schema/data/update [post]
func (h *handler) UpdateSchemaData() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.UpdateSchemaDataRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.inputService.UpdateSchemaData(req)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.UpdateSchemaDataFailed,
				code.Text(code.UpdateSchemaDataFailed)).WithError(err),
			)
			return
		}
		c.Payload("ok")
	}
}
