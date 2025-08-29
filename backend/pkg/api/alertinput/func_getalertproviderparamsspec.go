// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// GetAlertProviderParamsSpec Obtain alarm source parameter configuration
// @Summary Obtain alarm source parameter configuration
// @Description Obtain alarm source parameter configuration
// @Tags API.alertinput
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Request body request.GetAlertProviderParamsSpecRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.GetAlertProviderParamsSpecResponse
// @Failure 400 {object} code.Failure
// @Router /api/alertinput/source/paramspec [get]
func (h *handler) GetAlertProviderParamsSpec() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.GetAlertProviderParamsSpecRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				code.ParamBindError,
				err,
			)
			return
		}
		resp := h.inputService.GetProviderParamsSpec(c, req.SourceType)
		c.Payload(resp)
	}
}
