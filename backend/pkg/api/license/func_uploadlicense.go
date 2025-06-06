// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package license

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// UploadLicense uploads your license, sign public key, decrypt key.
// @Summary uploads your license, sign public key, decrypt key.
// @Description uploads your license, sign public key, decrypt key.
// @Tags API.license
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param output formData string true "output json"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} string "ok"
// @Failure 400 {object} code.Failure
// @Router /api/license/upload [post]
func (h *handler) UploadLicense() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UploadLicenseRequest)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		err := h.service.UploadLicense(req)
		if err != nil {
			c.HandleError(err, code.UploadLicenseError, nil)
			return
		}
		c.Payload("ok")
	}
}
