// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetSQLMetricsRequest struct {
	StartTime int64  `form:"startTime" binding:"min = 0"`                    // query start time
	EndTime   int64  `form:"endTime" binding:"required,gtfield = StartTime"` // query end time
	Service   string `form:"service" binding:"required"`                     // query service name
	Step      int64  `form:"step" binding:"min = 1000000"`                   // query step size (us)

	SortBy     string `form: "sortBy"` // sorting method,(latency,errorRate,tps) is sorted by latency by default
	*PageParam        // Paging Parameters
}
