// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type SetTTLRequest struct {
	DataType string `json:"dataType" binding:"required,oneof = logs trace k8s other topology"` // type (log, trace, Kubernetes, other)
	Day      int    `json:"day" binding:"required"`                                            // save data cycle days
}

type SetSingleTTLRequest struct {
	Name string `json:"name" binding:"required"` // table name
	Day  int    `json:"day" binding:"required"`  // save data cycle days
}
