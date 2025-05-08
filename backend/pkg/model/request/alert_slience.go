// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type RemoveAlertSlienceConfigRequest struct {
	AlertID string `json:"alertId" form:"alertId"`
}

type GetAlertSlienceConfigRequest struct {
	AlertID string `json:"alertId" form:"alertId"`
}

type SetAlertSlienceConfigRequest struct {
	AlertID     string `json:"alertId" form:"alertId"`
	ForDuration string `json:"forDuration" form:"forDuration"`
}
