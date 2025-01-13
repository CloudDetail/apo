// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

// CheckSchemaIsUsed 检查映射结构是否被使用
// @Summary 检查映射结构是否被使用
// @Description 检查映射结构是否被使用
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body alert.AlertSchemaRequest true "请求信息"
// @Success 200 {object} alert.CheckSchemaIsUsedReponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/schema/used/check [get]
func (h *handler) CheckSchemaIsUsed() core.HandlerFunc {
	return func(c core.Context) {
		req := new(alert.AlertSchemaRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

		alertSources, err := h.inputService.CheckSchemaIsUsed(req.Schema)
		if err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.CheckSchemaUsedFailed,
				code.Text(code.CheckSchemaUsedFailed)).WithError(err),
			)
			return
		}
		c.Payload(alert.CheckSchemaIsUsedReponse{
			IsUsing:          len(alertSources) > 0,
			AlertSourceNames: alertSources,
		})
	}
}
