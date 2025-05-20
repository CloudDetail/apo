// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// GetSubjectFeature Gets subject's feature permission.
// @Summary Gets subject's permission.
// @Description Gets subject's permission.
// @Tags API.permission
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param language query string false "language"
// @Param subjectId query int64 true "The id of authorized subject"
// @Param subjectType query string true "user, role, team"
// @Success 200 {object} response.GetSubjectFeatureResponse
// @Failure 400 {object} code.Failure
// @Router /api/permission/sub/feature [get]
func (h *handler) GetSubjectFeature() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetSubjectFeatureRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		resp, err := h.permissionService.GetSubjectFeature(c, req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.GetFeatureError,
				err,
			)
		}
		c.Payload(resp)
	}
}
