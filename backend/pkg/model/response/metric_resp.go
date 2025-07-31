// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import "github.com/CloudDetail/apo/backend/pkg/model"

type QueryPodsResponse struct {
	Pods []*model.Pod `json:"pods"`
}
