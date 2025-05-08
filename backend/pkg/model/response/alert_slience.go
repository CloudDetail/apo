// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"

type GetAlertSlienceConfigResponse struct {
	Slience *slienceconfig.AlertSlienceConfig `json:"slience"`
}

type ListAlertSlienceConfigResponse struct {
	Sliences []slienceconfig.AlertSlienceConfig `json:"sliences"`
}
