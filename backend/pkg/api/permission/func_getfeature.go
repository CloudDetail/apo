// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetFeature Gets all feature permission.
// @Summary Gets all feature permission.
// @Description Gets all feature permission.
// @Tags API.permission
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param language query string false "language"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetFeatureResponse
// @Failure 400 {object} code.Failure
// @Router /api/permission/feature [get]
func (h *handler) GetFeature() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetFeatureRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		if len(req.Language) == 0 {
			req.Language = model.TRANSLATION_ZH
		}

		resp, err := h.permissionService.GetFeature(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetFeatureError,
				err,
			)
			return
		}
		c.Payload(resp)
	}
}
