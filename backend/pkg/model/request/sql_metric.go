// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetSQLMetricsRequest struct {
	StartTime int64  `form:"startTime" binding:"min=0" json:"startTime"`                  // query start time
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time
	Service   string `form:"service" binding:"required" json:"service"`                   // query service name
	Step      int64  `form:"step" binding:"min=1000000" json:"step"`                      // query step size (us)

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
	GroupID    int64    `form:"groupId" json:"groupId"`

	SortBy     string `form:"sortBy" json:"sortBy"` // sorting method,(latency,errorRate,tps) is sorted by latency by default
	*PageParam        // Paging Parameters
}
