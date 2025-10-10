// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package app

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UploadLogo uploads logo
// @Summary uploads logo
// @Description uploads logo
// @Tags API.app
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param logoText formData string true "logoText"
// @Param logoImage formData file true "logoImage"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/app/upload [post]
func (h *handler) UploadLogo() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UploadLogoRequest)
		if err := c.ShouldBindMultiPartForm(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}

		err := h.service.UploadLogo(req)
		if err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.UploadLogoError,
				err,
			)
		}
		c.Payload("ok")
	}
}
