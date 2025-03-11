// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetMetricPQLRequest struct {
	Language string `json:"language" form:"language"`
}

type ListTargetTagsRequest struct {
	Language string `json:"language" form:"language"`
}
