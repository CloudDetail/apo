// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"errors"
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

// CheckAlertRule check whether the alarm rule name is available
// @Summary check whether the alarm rule name is available
// @Description check whether the alarm rule name is available
// @Tags API.alerts
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param alertRuleFile query string false "Query alarm rule file name"
// @Param group query string true "group name"
// @Param alert query string true "Alarm rule name"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.CheckAlertRuleResponse
// @Failure 400 {object} code.Failure
// @Router /api/alerts/rule/available  [get]
func (h *handler) CheckAlertRule() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.CheckAlertRuleRequest)
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				c.ErrMessage(code.ParamBindError)).WithError(err),
			)
			return
		}

		resp, err := h.alertService.CheckAlertRule(req)
		if err != nil {
			var vErr model.ErrWithMessage
			if errors.As(err, &vErr) {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					vErr.Code,
					c.ErrMessage(vErr.Code),
				).WithError(err),
				)
			} else {
				c.AbortWithError(core.Error(
					http.StatusBadRequest,
					code.AddAlertRuleError,
					c.ErrMessage(code.UpdateAlertRuleError),
				).WithError(err),
				)
			}
			return
		}

		c.Payload(resp)
	}
}
