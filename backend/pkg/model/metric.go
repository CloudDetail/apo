// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

type Pod struct {
	NodeName  string `json:"nodeName"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
}
