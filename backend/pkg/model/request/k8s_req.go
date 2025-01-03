// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetNamespaceInfoRequest struct {
	Namespace string `form:"namespace" binding:"required"`
}

type GetPodListRequest struct {
	Namespace string `form:"namespace" binding:"required"`
}

type GetPodInfoRequest struct {
	Namespace string `form:"namespace" binding:"required"`
	Pod       string `form:"pod" binding:"required"`
}
