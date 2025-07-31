// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type QueryPodsRequest struct {
	StartTime int64  `json:"startTime" binding:"min=0"`                    // query start time
	EndTime   int64  `json:"endTime" binding:"required,gtfield=StartTime"` // query end time
	NodeName  string `json:"nodeName"`                                     // query node name
	Namespace string `json:"namespace"`                                    // query namespace
	PodName   string `json:"pod"`                                          // query pod name
}
